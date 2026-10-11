// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package static

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

func TestStaticIsUpToDate(t *testing.T) {
	if runtime.GOOS == "android" {
		t.Skip("files not available on android")
	}
	oldBuf, err := ioutil.ReadFile("static.go")
	if err != nil {
		t.Errorf("error while reading static.go: %v\n", err)
	}

	newBuf, err := Generate()
	if err != nil {
		t.Errorf("error while generating static.go: %v\n", err)
	}

	// gen.go stamps time.Now().Year() into the leading "// Copyright" line,
	// so compare only what follows that first line (otherwise this fails in
	// every year after the one static.go was generated in).
	stripFirstLine := func(b []byte) []byte {
		if i := bytes.IndexByte(b, '\n'); i >= 0 {
			return b[i+1:]
		}
		return b
	}
	if bytes.Compare(stripFirstLine(oldBuf), stripFirstLine(newBuf)) != 0 {
		t.Error(`static.go is stale.  Run:
  $ go generate golang.org/x/tools/godoc/static
  $ git diff
to see the differences.`)

	}
}

// TestAppendQuote ensures that AppendQuote produces a valid literal.
func TestAppendQuote(t *testing.T) {
	var in, out bytes.Buffer
	for r := rune(0); r < unicode.MaxRune; r++ {
		in.WriteRune(r)
	}
	appendQuote(&out, in.Bytes())
	in2, err := strconv.Unquote(out.String())
	if err != nil {
		t.Fatalf("AppendQuote produced invalid string literal: %v", err)
	}
	if got, want := in2, in.String(); got != want {
		t.Fatal("AppendQuote modified string") // no point printing got/want: huge
	}
}

// TestJQueryXSSFixes checks that the jQuery 1.8.2 served by godoc (the
// generated Files["jquery.js"] entry, not just the jquery.js source file)
// carries the fixes for the known jQuery XSS and prototype-pollution
// vulnerabilities. Each case looks for the hardened construct in the
// minified source, or for the absence of the vulnerable one.
func TestJQueryXSSFixes(t *testing.T) {
	src, ok := Files["jquery.js"]
	if !ok {
		t.Fatal(`Files["jquery.js"] is missing`)
	}
	for _, tc := range []struct {
		name    string
		re      string // regular expression, or literal text when lit is set
		lit     bool
		present bool
	}{
		// CVE-2012-6708: jQuery(str) must only treat str as HTML when it
		// starts with '<', not when '<' appears anywhere in it.
		{"CVE-2012-6708/quickExpr-anchored", `/^(?:(<[\w\W]+>)[^>]*|#([\w\-]*))$/`, true, true},
		{"CVE-2012-6708/quickExpr-unanchored", `/^(?:[^#<]*(<[\w\W]+>)[^>]*$|#([\w\-]*)$)/`, true, false},
		// CVE-2015-9251: cross-domain responses without an explicit
		// dataType must not be evaluated as script.
		{"CVE-2015-9251/crossDomain-no-script", `\.ajaxPrefilter\(function\(\w+\)\{\w+\.crossDomain&&\(\w+\.contents\.script=!1\)\}\)`, false, true},
		// CVE-2019-11358: jQuery.extend(true, ...) must skip __proto__.
		{"CVE-2019-11358/extend-skips-proto", `if\(\w+==="__proto__"\|\|\w+===\w+\)continue;`, false, true},
		// CVE-2020-11022: self-closing tags must not be expanded to
		// "<tag></tag>" pairs before parsing.
		{"CVE-2020-11022/no-xhtml-tag-expansion", `"<$1></$2>"`, true, false},
		// CVE-2020-11023: <option> must not be wrapped in a <select>.
		{"CVE-2020-11023/no-option-wrap", `option:[1,"<select multiple='multiple'>","</select>"]`, true, false},
		// CVE-2020-7656: load() with a selector must parse the response
		// with parseHTML (which drops scripts) rather than a regexp that
		// misses "</script >".
		{"CVE-2020-7656/load-uses-parseHTML", `\("<div>"\)\.append\(\w+\.parseHTML\(\w+\)\)\.find\(`, false, true},
		{"CVE-2020-7656/load-no-regexp-strip", `\("<div>"\)\.append\(\w+\.replace\(`, false, false},
		// sonatype-2014-0026: parseHTML must default to an inert document
		// so inline event handlers do not run.
		{"sonatype-2014-0026/parseHTML-inert-document", `parseHTML:function\(\w+,\w+,\w+\)\{[^}]*\.implementation\.createHTMLDocument\(""\)`, false, true},
	} {
		var found bool
		if tc.lit {
			found = strings.Contains(src, tc.re)
		} else {
			found = regexp.MustCompile(tc.re).MatchString(src)
		}
		if found != tc.present {
			if tc.present {
				t.Errorf("%s: jquery.js does not contain %s", tc.name, tc.re)
			} else {
				t.Errorf("%s: jquery.js still contains %s", tc.name, tc.re)
			}
		}
	}
}
