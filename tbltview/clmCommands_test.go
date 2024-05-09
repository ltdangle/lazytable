package main

import (
	"testing"
)

func TestReplaceClmCommandRegex(t *testing.T) {
	testCases := []struct {
		Input           string
		ExpectedMatch   bool
		ExpectedSearch  string
		ExpectedReplace string
	}{
		{"replace 'foo' with 'bar'", true, "foo", "bar"},
		{"replace '' with 'bar'", true, "", "bar"},
		{"replace 'foo' with ''", true, "foo", ""},
		{"replace 'foo' with 'multi word replacement'", true, "foo", "multi word replacement"},
		{"replace 'foo' with 'multi word replacement'", true, "foo", "multi word replacement"},
		{"not a match", false, "", ""},
		{"replace 'unmatched with 'bar'", false, "", ""},
		{"incomplete 'foo' replacement", false, "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.Input, func(t *testing.T) {
			clm := NewReplaceClmCommand()
			ok, search, replace := clm.regex(tc.Input)

			if ok != tc.ExpectedMatch {
				t.Errorf("Expected match: %v, got: %v for input: %s",
					tc.ExpectedMatch, ok, tc.Input)
			}

			if ok {
				if search != tc.ExpectedSearch {
					t.Errorf("Expected search: %s, got: %s for input: %s",
						tc.ExpectedSearch, search, tc.Input)
				}

				if replace != tc.ExpectedReplace {
					t.Errorf("Expected replace: %s, got: %s for input: %s",
						tc.ExpectedReplace, replace, tc.Input)
				}
			}
		})
	}
}
