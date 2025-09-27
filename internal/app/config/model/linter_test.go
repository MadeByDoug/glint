package model

import (
	"encoding/json"
	"testing"
)

type predicateScenario struct {
	t     *testing.T
	name  string
	input string
	got   Predicate
	err   error
}

type predicateGiven struct{ s *predicateScenario }
type predicateWhen struct{ s *predicateScenario }
type predicateThen struct{ s *predicateScenario }

func givenPredicate(t *testing.T, name, input string) *predicateGiven {
	t.Helper()
	return &predicateGiven{s: &predicateScenario{t: t, name: name, input: input}}
}

func (g *predicateGiven) WhenUnmarshal() *predicateWhen {
	g.s.t.Helper()
	var p Predicate
	g.s.err = json.Unmarshal([]byte(g.s.input), &p)
	g.s.got = p
	return &predicateWhen{s: g.s}
}

func (w *predicateWhen) Then() *predicateThen {
	w.s.t.Helper()
	return &predicateThen{s: w.s}
}

func (th *predicateThen) ExpectValue(want Predicate) *predicateThen {
	th.s.t.Helper()
	if th.s.err != nil {
		th.s.t.Fatalf("%s: expected value %q, got error %v", th.s.name, want, th.s.err)
	}
	if th.s.got != want {
		th.s.t.Fatalf("%s: value = %q, want %q", th.s.name, th.s.got, want)
	}
	return th
}

func (th *predicateThen) ExpectError() *predicateThen {
	th.s.t.Helper()
	if th.s.err == nil {
		th.s.t.Fatalf("%s: expected error, got value %q", th.s.name, th.s.got)
	}
	return th
}

func (th *predicateThen) Done() {}

func TestPredicateUnmarshalJSON(t *testing.T) {
	givenPredicate(t, "kebab predicate", "\"kebab\"").
		WhenUnmarshal().Then().ExpectValue(PredicateKebab).Done()

	givenPredicate(t, "snake predicate", "\"snake\"").
		WhenUnmarshal().Then().ExpectValue(PredicateSnake).Done()

	givenPredicate(t, "camel predicate", "\"camel\"").
		WhenUnmarshal().Then().ExpectValue(PredicateCamel).Done()

	givenPredicate(t, "pascal predicate", "\"pascal\"").
		WhenUnmarshal().Then().ExpectValue(PredicatePascal).Done()

	givenPredicate(t, "lower predicate", "\"lower\"").
		WhenUnmarshal().Then().ExpectValue(PredicateLower).Done()

	givenPredicate(t, "upper predicate", "\"upper\"").
		WhenUnmarshal().Then().ExpectValue(PredicateUpper).Done()

	givenPredicate(t, "snake_case alias rejected", "\"snake_case\"").
		WhenUnmarshal().Then().ExpectError().Done()

	givenPredicate(t, "pascal-case alias rejected", "\"pascal-case\"").
		WhenUnmarshal().Then().ExpectError().Done()

	givenPredicate(t, "empty string rejected", "\"\"").
		WhenUnmarshal().Then().ExpectError().Done()

	givenPredicate(t, "unknown predicate rejected", "\"title-case\"").
		WhenUnmarshal().Then().ExpectError().Done()
}
