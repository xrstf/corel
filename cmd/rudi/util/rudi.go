// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package util

import (
	"fmt"

	"go.xrstf.de/rudi"
	"go.xrstf.de/rudi/cmd/rudi/types"
	"go.xrstf.de/rudi/pkg/coalescing"

	"go.xrstf.de/rudi-contrib/semver"
	"go.xrstf.de/rudi-contrib/uuid"
	"go.xrstf.de/rudi-contrib/yaml"
)

func SetupRudiContext(opts *types.Options, files []any) (rudi.Context, error) {
	var (
		document rudi.Document
		err      error
	)

	if len(files) > 0 {
		document, err = rudi.NewDocument(files[0])
		if err != nil {
			return rudi.Context{}, fmt.Errorf("cannot use first input as document: %w", err)
		}
	} else {
		document, _ = rudi.NewDocument(nil)
	}

	vars := rudi.NewVariables().
		Set("files", files)

	var coalescer coalescing.Coalescer
	switch opts.Coalescing {
	case "strict":
		coalescer = coalescing.NewStrict()
	case "pedantic":
		coalescer = coalescing.NewPedantic()
	case "humane":
		coalescer = coalescing.NewHumane()
	default:
		return rudi.Context{}, fmt.Errorf("unknown coalescing mode %q", opts.Coalescing)
	}

	funcs := rudi.NewBuiltInFunctions().
		Add(semver.Functions).
		Add(uuid.Functions).
		Add(yaml.Functions)

	ctx := rudi.NewContext(document, vars, funcs, coalescer)

	return ctx, nil
}
