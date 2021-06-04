package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func doActionTest(ctx *context) error {
	if ctx.GetVariable("HK") == nil {
		ctx.SetVariable("HK", 0)
	}

	ctx.SetVariable("HK", ctx.GetVariable("HK").(int)+1)
	return nil
}

func undoActionTest(ctx context) {

}

func execTestRoute(route *State) *routeRunner {
	rh := newRouteRunner(route, nil)
	ctx, _ := NewContext()

	rh.run(ctx, nil)
	return rh
}

func TestDefineUnconditionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 3, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))

	rf := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh = execTestRoute(rf.GetStartState())

	assert.Equal(t, 2, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineNestedConditionalRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 6, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		Otherwise().
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return false }).
		AddNextStep(doActionTest, undoActionTest).
		Otherwise().
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineRoute(t *testing.T) {
	r := NewTransactionalRoute("TEST_ROUTE").
		AddNextStep(doActionTest, undoActionTest).
		When(func(ctx context) bool { return true }).
		AddNextStep(doActionTest, undoActionTest).
		AddNextStep(doActionTest, undoActionTest).
		Otherwise().
		AddNextStep(doActionTest, undoActionTest).
		End().
		AddNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.GetStartState())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}
