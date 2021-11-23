package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefineUnconditionalRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		AddNextStep("3", doActionTest)

	rr := execTestRoute(r.GetStartState())

	assert.Equal(t, 3, rr.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionalRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest).
		AddNextStep("when_2", doActionTest).
		AddNextStep("when_3", doActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))

	rf := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep("when_1", doActionTest).
		AddNextStep("when_2", doActionTest).
		AddNextStep("when_3", doActionTest)

	rh = execTestRoute(rf.GetStartState())

	assert.Equal(t, 2, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineNestedConditionalRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_when_1", doActionTest).
		AddNextStep("when_when_2", doActionTest).
		AddNextStep("when_when_3", doActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 6, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", doActionTest).
		AddNextStep("when_2", doActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest).
		AddNextStep("otherwise_2", doActionTest).
		AddNextStep("otherwise_3", doActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		AddNextStep("2", doActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep("condition_1", doActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest).
		AddNextStep("otherwise_2", doActionTest).
		AddNextStep("otherwise_3", doActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineRoute(t *testing.T) {
	r := NewNonTransactionalRoute("TEST_ROUTE").
		AddNextStep("1", doActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep("condition_1", doActionTest).
		AddNextStep("condition_2", doActionTest).
		Otherwise().
		AddNextStep("otherwise_1", doActionTest).
		End().
		AddNextStep("2", doActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}
