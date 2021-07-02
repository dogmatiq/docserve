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
	"github.com/dogmatiq/dapper"
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
	logging.Log(a.Logger, "[%s] analyzing default branch (%s)", r.GetFullName(), r.GetDefaultBranch())

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
	sha := branch.GetCommit().GetSHA()

	logging.Log(a.Logger, "[%s] head commit is %s", r.GetFullName(), sha)

	ok, err := a.mayContainApplications(
		ctx,
		c,
		r,
		sha,
	)
	if !ok || err != nil {
		return err
	}

	pkgs, err := a.loadPackages(ctx, c, r, sha)
	if err != nil {
		return err
	}

	apps := a.analyzePackages(r, pkgs)

	return a.sync(
		ctx,
		r,
		sha,
		apps,
	)
}

func (a *Analyzer) mayContainApplications(
	ctx context.Context,
	c *github.Client,
	r *github.Repository,
	sha string,
) (bool, error) {
	content, _, res, err := c.Repositories.GetContents(
		ctx,
		r.GetOwner().GetLogin(),
		r.GetName(),
		"go.mod",
		&github.RepositoryContentGetOptions{
			Ref: sha,
		},
	)
	if err != nil {
		if res.StatusCode == http.StatusNotFound {
			logging.Log(a.Logger, "[%s] skipping analysis, go.mod not found", r.GetFullName())
			return false, nil
		}

		return false, err
	}

	data, err := content.GetContent()
	if err != nil {
		dapper.Print(err)
		return false, err
	}

	mod, err := modfile.ParseLax(
		"go.mod",
		[]byte(data),
		nil,
	)
	if err != nil {
		logging.Log(a.Logger, "[%s] skipping analysis, go.mod is invalid: %s", r.GetFullName(), err)
		return false, nil
	}

	for _, req := range mod.Require {
		if req.Mod.Path == "github.com/dogmatiq/dogma" {
			return true, nil
		}
	}

	logging.Log(a.Logger, "[%s] skipping analysis, module does not depend on github.com/dogmatiq/dogma", r.GetFullName())

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
	valid := make([]*packages.Package, 0, len(pkgs))

	for _, pkg := range pkgs {
		if len(pkg.Errors) == 0 {
			logging.Log(a.Logger, "[%s] analyzing package: %s", r.GetFullName(), pkg.PkgPath)
			valid = append(valid, pkg)
		} else {
			for _, err := range pkg.Errors {
				logging.Log(a.Logger, "[%s] unable to analyze package %s: %s", r.GetFullName(), pkg.PkgPath, err)
			}
		}
	}

	apps := static.FromPackages(valid)

	for _, app := range apps {
		logging.Log(a.Logger, "[%s] discovered dogma application: %s", r.GetFullName(), app.Identity())
	}

	return apps
}

func (a *Analyzer) sync(
	ctx context.Context,
	r *github.Repository,
	sha string,
	apps []configkit.Application,
) error {
	tx, err := a.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	if err := persistence.SyncRepository(
		ctx,
		tx,
		r,
		sha,
	); err != nil {
		return err
	}

	for _, app := range apps {
		if err := persistence.SyncApplication(
			ctx,
			tx,
			r,
			app,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
