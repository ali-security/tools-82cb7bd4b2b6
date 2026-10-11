// create_mod builds the golang.org/x/tools module zip with
// golang.org/x/mod/zip.CreateFromDir (the same inclusion rules the module
// proxy applies). It lives under .github so `go build ./...` ignores it; the
// workflow builds it in a separate temp module pinned to x/mod v0.24.0.
//
// Usage: create_mod <module-path> <version> <source-dir> <output-zip>
package main

import (
	"log"
	"os"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatal("usage: create_mod <module-path> <version> <source-dir> <output-zip>")
	}
	mv := module.Version{Path: os.Args[1], Version: os.Args[2]}
	f, err := os.Create(os.Args[4])
	if err != nil {
		log.Fatal(err)
	}
	if err := zip.CreateFromDir(f, mv, os.Args[3]); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
