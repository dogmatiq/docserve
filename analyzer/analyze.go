package analyzer

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/static"
	"github.com/dogmatiq/docserve/githubx"
	"github.com/dogmatiq/docserve/persistence"
	"github.com/dogmatiq/dodeca/logging"
	"github.com/google/go-github/v35/github"
	"go.uber.org/multierr"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type Analyzer struct {
	DB           *sql.DB
	GitHubClient *github.Client
	HTTPClient   *http.Client
	Logger       logging.Logger
}

func (a *Analyzer) Analyze(ctx context.Context, r *github.Repository) error {
	if err := a.analyze(ctx, r); err != nil {
		return fmt.Errorf("unable to analyze %s: %w", r.GetFullName(), err)
	}

	return nil
}

func (a *Analyzer) analyze(ctx context.Context, r *github.Repository) error {
	c, err := githubx.NewClientForRepository(ctx, a.GitHubClient, r)
	if err != nil {
		return err
	}

	branch, _, err := c.Repositories.GetBranch(
		ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		r.GetDefaultBranch(),
	)
	if err != nil {
		return err
	}

	commitHash := branch.GetCommit().GetSHA()

	needsSync, err := persistence.RepositoryNeedsSync(
		ctx,
		a.DB,
		r,
		commitHash,
	)
	if err != nil {
		return err
	}

	if !needsSync {
		logging.Log(
			a.Logger,
			"[%s] skipping analysis of %s branch (%s), commit has already been analysed",
			r.GetFullName(),
			r.GetDefaultBranch(),
			commitHash,
		)
		return nil
	}

	ok, err := a.mayContainApplications(
		ctx,
		c,
		r,
		commitHash,
	)
	if err != nil {
		return err
	}

	var apps []configkit.Application

	if ok {
		pkgs, err := a.loadPackages(ctx, c, r, commitHash)
		if err != nil {
			return err
		}

		apps = a.analyzePackages(r, pkgs)
	}

	return persistence.SyncRepository(
		ctx,
		a.DB,
		r,
		commitHash,
		apps,
	)
}

func (a *Analyzer) mayContainApplications(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
	commitHash string,
) (bool, error) {
	content, _, res, err := c.Repositories.GetContents(
		ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		"go.mod",
		&github.RepositoryContentGetOptions{
			Ref: commitHash,
		},
	)
	if err != nil {
		if res.StatusCode == http.StatusNotFound {
			logging.Log(
				a.Logger,
				"[%s] skipping analysis of %s branch (%s), go.mod file not present",
				r.GetFullName(),
				r.GetDefaultBranch(),
				commitHash,
			)

			return false, nil
		}

		return false, err
	}

	data, err := content.GetContent()
	if err != nil {
		return false, err
	}

	mod, err := modfile.ParseLax(
		"go.mod",
		[]byte(data),
		nil,
	)
	if err != nil {
		logging.Log(
			a.Logger,
			"[%s] skipping analysis of %s branch (%s), go.mod file is invalid: %s",
			r.GetFullName(),
			r.GetDefaultBranch(),
			commitHash,
			err,
		)

		return false, nil
	}

	for _, req := range mod.Require {
		if req.Mod.Path == "github.com/dogmatiq/dogma" {
			logging.Log(
				a.Logger,
				"[%s] analyzing %s branch (%s), dogma version %s",
				r.GetFullName(),
				r.GetDefaultBranch(),
				commitHash,
				req.Mod.Version,
			)

			return true, nil
		}
	}

	logging.Log(
		a.Logger,
		"[%s] skipping analysis of %s branch (%s), go.mod file does not specify github.com/dogmatiq/dogma as a requirement",
		r.GetFullName(),
		r.GetDefaultBranch(),
		commitHash,
	)

	return false, nil
}

func (a *Analyzer) loadPackages(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
	sha string,
) (_ []*packages.Package, err error) {
	dir, err := ioutil.TempDir(
		"",
		fmt.Sprintf(
			"dogma-docserve_%s-%s_%s",
			r.GetOwner().GetLogin(),
			r.GetName(),
			sha,
		),
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Remove the source immediately after we load the packages from it so
		// that it doesn't spend any longer on disk than it needs to.
		//
		// Fail hard if we can't remove it, as there are potential security
		// implications.
		err = multierr.Append(
			err,
			os.RemoveAll(dir),
		)
	}()

	if err := a.download(ctx, c, r, sha, dir); err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Context: ctx,
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedDeps,
		Dir: dir,
	}

	return packages.Load(cfg, "./...")
}

func (a *Analyzer) download(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
	sha, dir string,
) error {
	logging.Log(a.Logger, "[%s] downloading source archive to %s", r.GetFullName(), dir)

	url, _, err := c.Repositories.GetArchiveLink(ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		github.Tarball,
		&github.RepositoryContentGetOptions{
			Ref: sha,
		},
		true,
	)
	if err != nil {
		return err
	}

	return downloadAndUncompress(
		ctx,
		url.String(),
		dir,
		a.HTTPClient,
	)
}

func (a *Analyzer) analyzePackages(
	r *github.Repository,
	pkgs []*packages.Package,
) []configkit.Application {
	n := len(pkgs)

	for i, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			logging.Log(
				a.Logger,
				"[%s] analyzing %s",
				r.GetFullName(),
				pkg.PkgPath,
			)
		} else {
			for _, err := range pkg.Errors {
				logging.Log(
					a.Logger,
					"[%s] unable to analyze %s: %s",
					r.GetFullName(),
					pkg.PkgPath,
					err,
				)
			}

			pkgs[i], pkgs[n-1] = pkgs[n-1], pkgs[i]
			n--
		}
	}

	apps := static.FromPackages(pkgs[:n])

	for _, app := range apps {
		logging.Log(
			a.Logger,
			"[%s] discovered %s application",
			r.GetFullName(),
			app.Identity(),
		)
	}

	return apps
}
