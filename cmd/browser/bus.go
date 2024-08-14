package main

import (
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/minibus"
)

func init() {
	imbue.With0(
		container,
		func(ctx imbue.Context) ([]minibus.Func, error) {
			return []minibus.Func{}, nil
		},
	)
}
