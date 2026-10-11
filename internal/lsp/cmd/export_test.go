// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	ExampleLine   = exampleLine
	ExampleColumn = exampleColumn
	ExampleOffset = exampleOffset
)

// TestServeRejectsImplicitAllInterfaces checks that "gopls serve -listen=:port"
// refuses to start a server that would implicitly bind all network interfaces.
func TestServeRejectsImplicitAllInterfaces(t *testing.T) {
	app := New("", nil)
	for _, addr := range []string{":0", ":12345"} {
		s := &Serve{Address: addr, app: app}
		errc := make(chan error, 1)
		go func() { errc <- s.Run(context.Background()) }()
		select {
		case err := <-errc:
			if err == nil {
				t.Fatalf("Serve{Address: %q}.Run succeeded, want error", addr)
			}
			if !strings.Contains(err.Error(), "implicitly binds all network interfaces") {
				t.Errorf("Serve{Address: %q}.Run returned %q, want implicit all-interfaces error", addr, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatalf("Serve{Address: %q}.Run is listening on all network interfaces, want it rejected", addr)
		}
	}
}

// TestServeHasNoPortFlag checks that the -port flag, which listened on
// ":port" (all network interfaces), is no longer offered.
func TestServeHasNoPortFlag(t *testing.T) {
	typ := reflect.TypeOf(Serve{})
	for i := 0; i < typ.NumField(); i++ {
		if name := typ.Field(i).Tag.Get("flag"); name == "port" {
			t.Errorf("Serve has a -port flag (field %s), want none", typ.Field(i).Name)
		}
	}
}
