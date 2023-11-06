// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.18 && !go1.22
// +build go1.18,!go1.22

package buildinfo

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/vuln/internal/test"
)

// TestStrippedBinary checks that there is no symbol table for
// stripped binaries. This does not include darwin binaries.
// For more info, see #61051.
func TestStrippedBinary(t *testing.T) {
	// We exclude darwin as its stripped binaries seem to
	// preserve the symbol table. See TestStrippedDarwin.
	testAll(t, []string{"linux", "windows", "freebsd"}, []string{"amd64", "386", "arm", "arm64"},
		func(t *testing.T, goos, goarch string) {
			binary, done := test.GoBuild(t, "testdata", "", true, "GOOS", goos, "GOARCH", goarch)
			defer done()

			f, err := os.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			_, syms, _, err := ExtractPackagesAndSymbols(f)
			if err != nil {
				t.Fatal(err)
			}
			if syms != nil {
				t.Errorf("want empty symbol table; got %v symbols", len(syms))
			}
		})
}

// TestStrippedDarwin checks that the symbol table exists and
// is complete on darwin even in the presence of stripping.
// For more info, see #61051.
func TestStrippedDarwin(t *testing.T) {
	testAll(t, []string{"darwin"}, []string{"amd64", "386"},
		func(t *testing.T, goos, goarch string) {
			binary, done := test.GoBuild(t, "testdata", "", true, "GOOS", goos, "GOARCH", goarch)
			defer done()

			f, err := os.Open(binary)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			_, syms, _, err := ExtractPackagesAndSymbols(f)
			if err != nil {
				t.Fatal(err)
			}

			got := mainSortedSymbols(syms)
			want := []Symbol{
				{"main", "f"},
				{"main", "g"},
				{"main", "main"},
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("(-want,+got):%s", diff)
			}
		})
}
