package main

import (
	"github.com/dogmatiq/browser/components/correlator"
	"github.com/dogmatiq/imbue"
	"github.com/dogmatiq/minibus"
)

func init() {
	imbue.With0(
		container,
		func(
			ctx imbue.Context,
		) (*correlator.Correlator, error) {
			return &correlator.Correlator{}, nil
		},
	)

	imbue.Decorate1(
		container,
		func(
			ctx imbue.Context,
			funcs []minibus.Func,
			c *correlator.Correlator,
		) ([]minibus.Func, error) {
			return append(funcs, c.Run), nil
		},
	)
}
