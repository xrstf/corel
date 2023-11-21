// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package builtin

import (
	"errors"
	"fmt"

	"go.xrstf.de/rudi/pkg/eval"
	"go.xrstf.de/rudi/pkg/eval/coalescing"
	"go.xrstf.de/rudi/pkg/eval/types"
	"go.xrstf.de/rudi/pkg/lang/ast"
	"go.xrstf.de/rudi/pkg/pathexpr"
)

// (if COND:Expr YES:Expr NO:Expr?)
func ifFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size < 2 || size > 3 {
		return nil, fmt.Errorf("expected 2 or 3 arguments, got %d", size)
	}

	_, condition, err := eval.EvalExpression(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("condition: %w", err)
	}

	success, ok := condition.(ast.Bool)
	if !ok {
		return nil, fmt.Errorf("condition is not bool, but %T", condition)
	}

	if success {
		// discard context changes from the true path
		_, result, err := eval.EvalExpression(ctx, args[1])
		return result, err
	}

	// optional else part
	if len(args) > 2 {
		// discard context changes from the false path
		_, result, err := eval.EvalExpression(ctx, args[2])
		return result, err
	}

	return ast.Null{}, nil
}

// (do STEP:Expr+)
func doFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size < 1 {
		return nil, fmt.Errorf("expected 1+ arguments, got %d", size)
	}

	tupleCtx := ctx

	var (
		result any
		err    error
	)

	// do not use evalArgs(), as we want to inherit the context between expressions
	for i, arg := range args {
		tupleCtx, result, err = eval.EvalExpression(tupleCtx, arg)
		if err != nil {
			return nil, fmt.Errorf("argument #%d: %w", i, err)
		}
	}

	return result, nil
}

// (has? SYM:SymbolWithPathExpression)
func hasFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size != 1 {
		return nil, fmt.Errorf("expected 1 argument, got %d", size)
	}

	var (
		expr     ast.Expression
		pathExpr *ast.PathExpression
	)

	// separate base value expression from the path expression

	if symbol, ok := args[0].(ast.Symbol); ok {
		pathExpr = symbol.PathExpression

		if symbol.Variable != nil {
			symbol.PathExpression = nil
		} else {
			// for bare path expressions
			symbol.PathExpression = &ast.PathExpression{}
		}

		expr = symbol
	}

	if vectorNode, ok := args[0].(ast.VectorNode); ok {
		pathExpr = vectorNode.PathExpression
		vectorNode.PathExpression = nil
		expr = vectorNode
	}

	if objectNode, ok := args[0].(ast.ObjectNode); ok {
		pathExpr = objectNode.PathExpression
		objectNode.PathExpression = nil
		expr = objectNode
	}

	if tuple, ok := args[0].(ast.Tuple); ok {
		pathExpr = tuple.PathExpression
		tuple.PathExpression = nil
		expr = tuple
	}

	if expr == nil {
		return nil, fmt.Errorf("expected Symbol, Vector, Object or Tuple, got %T", args[0])
	}

	if pathExpr == nil {
		return nil, errors.New("argument has no path expression")
	}

	// pre-evaluate the path
	evaluatedPath, err := eval.EvalPathExpression(ctx, pathExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid path expression: %w", err)
	}

	// evaluate the base value
	_, value, err := eval.EvalExpression(ctx, expr)
	if err != nil {
		return nil, err
	}

	_, err = eval.TraverseEvaluatedPathExpression(ctx, value, *evaluatedPath)
	if err != nil {
		return ast.Bool(false), nil
	}

	return ast.Bool(true), nil
}

// (default TEST:Expression FALLBACK:any)
func defaultFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", size)
	}

	_, result, err := eval.EvalExpression(ctx, args[0])
	if err != nil {
		return nil, fmt.Errorf("argument #0: %w", err)
	}

	isEmpty, err := coalescing.IsEmpty(result)
	if err != nil {
		return nil, fmt.Errorf("argument #0: %w", err)
	}

	if !isEmpty {
		return result, nil
	}

	_, result, err = eval.EvalExpression(ctx, args[1])
	if err != nil {
		return nil, fmt.Errorf("argument #1: %w", err)
	}

	return result, nil
}

// (try TEST:Expression FALLBACK:any?)
func tryFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size < 1 || size > 2 {
		return nil, fmt.Errorf("expected 1 or 2 arguments, got %d", size)
	}

	_, result, err := eval.EvalExpression(ctx, args[0])
	if err != nil {
		if len(args) == 1 {
			return nil, nil
		}

		_, result, err = eval.EvalExpression(ctx, args[1])
		if err != nil {
			return nil, fmt.Errorf("argument #1: %w", err)
		}
	}

	return result, nil
}

// (set VAR:Variable VALUE:any)
// (set EXPR:PathExpression VALUE:any)
func setFunction(ctx types.Context, args []ast.Expression) (types.Context, any, error) {
	if size := len(args); size != 2 {
		return ctx, nil, fmt.Errorf("expected 2 arguments, got %d", size)
	}

	symbol, ok := args[0].(ast.Symbol)
	if !ok {
		return ctx, nil, fmt.Errorf("argument #0 is not a symbol, but %T", args[0])
	}

	// catch symbols that are technically invalid
	if symbol.Variable == nil && symbol.PathExpression == nil {
		return ctx, nil, fmt.Errorf("argument #0: must be path expression or variable, got %s", symbol.ExpressionName())
	}

	// discard any context changes within the newValue expression
	_, newValue, err := eval.EvalExpression(ctx, args[1])
	if err != nil {
		return ctx, nil, fmt.Errorf("argument #1: %w", err)
	}

	// pre-evaluate the path
	var pathExpr *ast.EvaluatedPathExpression
	if p := symbol.PathExpression; p != nil {
		pathExpr, err = eval.EvalPathExpression(ctx, p)
		if err != nil {
			return ctx, nil, fmt.Errorf("argument #1: invalid path expression: %w", err)
		}
	}

	// get the current value
	var currentValue any

	if symbol.Variable != nil {
		varName := string(*symbol.Variable)

		// a non-existing variable is fine, this is how you define new variables in the first place
		currentValue, _ = ctx.GetVariable(varName)
	} else {
		doc := ctx.GetDocument()
		currentValue = doc.Get()
	}

	// if there is a path expression, merge in the new value
	updatedValue := newValue
	if pathExpr != nil {
		updatedValue, err = pathexpr.Set(currentValue, pathexpr.FromEvaluatedPath(*pathExpr), newValue)
		if err != nil {
			return ctx, nil, fmt.Errorf("cannot set value in %T at %s: %w", currentValue, pathExpr, err)
		}
	}

	// set a variable, which will result in a new context
	if symbol.Variable != nil {
		varName := string(*symbol.Variable)

		// make the variable's value the return value, so `(def $foo 12)` = 12
		return ctx.WithVariable(varName, updatedValue), newValue, nil
	}

	// update the global document
	// (the document Go struct stays the same, so this does not result in a new context)
	doc := ctx.GetDocument()
	doc.Set(updatedValue)

	return ctx, newValue, nil
}

// (empty? VALUE:any)
func isEmptyFunction(ctx types.Context, args []ast.Expression) (any, error) {
	if size := len(args); size != 1 {
		return nil, fmt.Errorf("expected 1 or 2 arguments, got %d", size)
	}

	_, result, err := eval.EvalExpression(ctx, args[0])
	if err != nil {
		return nil, err
	}

	empty, err := coalescing.IsEmpty(result)
	if err != nil {
		return nil, err
	}

	return ast.Bool(empty), nil
}
