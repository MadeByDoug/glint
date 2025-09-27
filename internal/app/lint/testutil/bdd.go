// internal/app/lint/testutil/bdd.go
package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/MrBigCode/glint/internal/app/infra/output/reporting"
	"github.com/MrBigCode/glint/internal/app/lint"
)

// Expect is a simple matcher for Then-step assertions.
// Msg is substring-matched to keep scenarios concise.
type Expect struct {
	Code string
	Msg  string
}

// Scenario holds the evolving state across Given/When/Then.
type Scenario struct {
	t      *testing.T
	name   string
	tree   *lint.Tree
	checks []lint.Checker
	diags  []reporting.Report
}

// --- Given ---

type GivenStep struct{ s *Scenario }

// Given starts a new scenario with a readable name.
func Given(t *testing.T, name string) *GivenStep {
	t.Helper()
	return &GivenStep{s: &Scenario{t: t, name: name}}
}

// Tree builds an in-memory filesystem using your builder.
func (g *GivenStep) Tree(build func(*B)) *GivenStep {
	b := New()
	if build != nil {
		build(b)
	}
	g.s.tree = b.Build()
	return g
}

// Checks registers the checkers to run during When.
func (g *GivenStep) Checks(checks ...lint.Checker) *GivenStep {
	g.s.checks = append(g.s.checks, checks...)
	return g
}

// --- When ---

type WhenStep struct{ s *Scenario }

// WhenLint runs the pure lint orchestration.
func (g *GivenStep) WhenLint(ctx context.Context) *WhenStep {
	g.s.t.Helper()
	g.s.diags = lint.Lint(ctx, g.s.tree, g.s.checks...)
	return &WhenStep{s: g.s}
}

// --- Then ---

type ThenStep struct{ s *Scenario }

func (w *WhenStep) Then() *ThenStep {
	w.s.t.Helper()
	return &ThenStep{s: w.s}
}

// ExpectNone asserts there are no diagnostics.
func (th *ThenStep) ExpectNone() *ThenStep {
	th.s.t.Helper()
	if len(th.s.diags) != 0 {
		th.s.t.Fatalf("%s: expected no diagnostics, got %d\n%+v", th.s.name, len(th.s.diags), th.s.diags)
	}
	return th
}

// ExpectContains asserts that each expected (Code, substring of Msg) exists.
func (th *ThenStep) ExpectContains(expects ...Expect) *ThenStep {
	th.s.t.Helper()
	for _, e := range expects {
		if th.containsExpectation(e) {
			continue
		}
		th.failMissingExpectation(e)
	}
	return th
}

func (th *ThenStep) containsExpectation(e Expect) bool {
	for _, d := range th.s.diags {
		if d.Code != e.Code {
			continue
		}
		if e.Msg == "" || strings.Contains(d.Msg, e.Msg) {
			return true
		}
	}
	return false
}

func (th *ThenStep) failMissingExpectation(e Expect) {
	th.s.t.Fatalf("%s: missing expected diag code=%s msg~%q\nGOT: %+v", th.s.name, e.Code, e.Msg, th.s.diags)
}

// ExpectOnlyCodes asserts that the set of diagnostic codes equals the provided set.
func (th *ThenStep) ExpectOnlyCodes(codes ...string) *ThenStep {
	th.s.t.Helper()
	got := map[string]int{}
	for _, d := range th.s.diags {
		got[d.Code]++
	}
	want := map[string]int{}
	for _, c := range codes {
		want[c]++
	}
	if len(got) != len(want) {
		th.s.t.Fatalf("%s: expected codes %v, got %v", th.s.name, want, got)
	}
	for c := range want {
		if _, ok := got[c]; !ok {
			th.s.t.Fatalf("%s: expected code %s not present; got %v", th.s.name, c, got)
		}
	}
	return th
}

func (th *ThenStep) ExpectCount(n int) *ThenStep {
	th.s.t.Helper()
	if len(th.s.diags) != n {
		th.s.t.Fatalf("%s: expected %d diagnostics, got %d\n%+v", th.s.name, n, len(th.s.diags), th.s.diags)
	}
	return th
}

func (th *ThenStep) ExpectHasSeverity(sev reporting.Severity) *ThenStep {
	th.s.t.Helper()
	for _, d := range th.s.diags {
		if d.Severity == sev {
			return th
		}
	}
	th.s.t.Fatalf("%s: expected at least one diagnostic with severity=%s; got=%+v", th.s.name, sev, th.s.diags)
	return th
}

// Done ends the chain (no-op but keeps call sites explicit).
func (th *ThenStep) Done() {}
