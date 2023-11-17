// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package eval

import (
	"fmt"

	"go.xrstf.de/otto/pkg/lang/ast"
	"go.xrstf.de/otto/pkg/lang/eval/types"
)

func EvalIdentifier(ctx types.Context, ident ast.Identifier) (types.Context, any, error) {
	return ctx, nil, fmt.Errorf("unexpected identifier: %v", ident)
}
