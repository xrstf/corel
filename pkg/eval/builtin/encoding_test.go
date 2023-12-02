// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package builtin

import (
	"testing"

	"go.xrstf.de/rudi/pkg/testutil"
)

func TestToBase64Function(t *testing.T) {
	testcases := []testutil.Testcase{
		{
			Expression: `(to-base64)`,
			Invalid:    true,
		},
		{
			Expression: `(to-base64 "too" "many")`,
			Invalid:    true,
		},
		{
			Expression: `(to-base64 true)`,
			Invalid:    true,
		},
		{
			Expression: `(to-base64 1)`,
			Invalid:    true,
		},
		{
			// strict coalescing allows null to turn into ""
			Expression: `(to-base64 null)`,
			Expected:   "",
		},
		{
			Expression: `(to-base64 "")`,
			Expected:   "",
		},
		{
			Expression: `(to-base64 " ")`,
			Expected:   "IA==",
		},
		{
			Expression: `(to-base64 (concat "" "f" "o" "o"))`,
			Expected:   "Zm9v",
		},
		{
			Expression: `(to-base64 "test")`,
			Expected:   "dGVzdA==",
		},
	}

	for _, testcase := range testcases {
		testcase.Functions = AllFunctions
		t.Run(testcase.String(), testcase.Run)
	}
}

func TestFromBase64Function(t *testing.T) {
	testcases := []testutil.Testcase{
		{
			Expression: `(from-base64)`,
			Invalid:    true,
		},
		{
			Expression: `(from-base64 "too" "many")`,
			Invalid:    true,
		},
		{
			Expression: `(from-base64 true)`,
			Invalid:    true,
		},
		{
			Expression: `(from-base64 1)`,
			Invalid:    true,
		},
		{
			Expression: `(from-base64 "definitely-not-base64")`,
			Invalid:    true,
		},
		{
			// strict coalescing allows null to turn into ""
			Expression: `(from-base64 null)`,
			Expected:   "",
		},
		{
			// should be able to recover
			Expression: `(try (from-base64 "definitely-not-base64") "fallback")`,
			Expected:   "fallback",
		},
		{
			Expression: `(from-base64 "")`,
			Expected:   "",
		},
		{
			Expression: `(from-base64 "IA==")`,
			Expected:   " ",
		},
		{
			Expression: `(from-base64 (concat "" "Z" "m" "9" "v"))`,
			Expected:   "foo",
		},
		{
			Expression: `(from-base64 "dGVzdA==")`,
			Expected:   "test",
		},
	}

	for _, testcase := range testcases {
		testcase.Functions = AllFunctions
		t.Run(testcase.String(), testcase.Run)
	}
}
