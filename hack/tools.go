//go:build tools

// Package tools pins nogo analyzer dependencies in go.mod/go.sum so `go mod
// tidy` doesn't drop them. None of these are imported by the built binary.
package tools

import (
	_ "github.com/gordonklaus/ineffassign/pkg/ineffassign"
	_ "github.com/kisielk/errcheck/errcheck"
	_ "github.com/nunnatsa/ginkgolinter"
	_ "honnef.co/go/tools/staticcheck"
	_ "honnef.co/go/tools/unused"
)
