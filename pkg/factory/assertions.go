package factory

import (
	"testing"
)

type composedAssertion[T kubeObjects, testcase testcaseType] struct {
	t          *testing.T
	predicates []predicate[T, testcase]
}

func (a *composedAssertion[T, K]) assertTest(Obj T, testcase K) {
	for _, pred := range a.predicates {
		if !pred.Fn(Obj, testcase) {
			a.t.Errorf("Assertion %s failed", pred.Name)
		}
	}
}

func ComposeAssertions[T kubeObjects, K testcaseType](predicates []predicate[T, K], t *testing.T) *composedAssertion[T, K] {
	return &composedAssertion[T, K]{
		predicates: predicates,
		t:          t,
	}
}
