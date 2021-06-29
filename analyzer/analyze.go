package analyzer

import (
	"context"
	"path/filepath"

	"github.com/dogmatiq/configkit"
	"github.com/dogmatiq/configkit/static"
	"github.com/dogmatiq/dodeca/logging"
	"golang.org/x/tools/go/packages"
	"gorm.io/gorm"
)

type Analyzer struct {
	DB     *gorm.DB
	Logger logging.Logger
}

func (a *Analyzer) Analyze(
	ctx context.Context,
	path string,
) error {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.LoadAllSyntax,
		Dir:     path,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			logging.Log(a.Logger, "unable to analyze %s: %s", pkg.PkgPath, err)
		}
	}

	return nil
}

func loadPackages(ctx context.Context, path string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.LoadAllSyntax,
		Dir:     path,
	}

	return packages.Load(cfg, "./...")
}

// loadConfigsFromPackages returns the configuration for all applications
// defined within the packages that match the given patterns.
func loadConfigsFromPackages(
	ctx context.Context,
	patterns []string,
) ([]configkit.Application, error) {
	var applications []configkit.Application

	for _, pattern := range patterns {
		cfg := packages.Config{
			Context: ctx,
			Mode:    packages.LoadAllSyntax,
			Dir:     pattern,
		}

		if filepath.Base(pattern) == "..." {
			cfg.Dir = filepath.Dir(pattern)
			pattern = "./..."
		} else {
			pattern = "."
		}

		pkgs, err := packages.Load(&cfg, pattern)
		if err != nil {
			return nil, err
		}

		for _, pkg := range pkgs {
			for _, err := range pkg.Errors {
				return nil, err
			}
		}

		applications = append(
			applications,
			static.FromPackages(pkgs)...,
		)

	}

	return applications, nil
}
