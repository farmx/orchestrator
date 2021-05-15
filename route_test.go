package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type alwaysPassTransactionMock struct {
	TransactionalStep
}

func doActionTest(ctx *context) error {
	if ctx.GetVariable("HK") == nil {
		ctx.SetVariable("HK", 0)
	}

	ctx.SetVariable("HK", ctx.GetVariable("HK").(int)+1)
	return nil
}

func undoActionTest(ctx context) {

}

func execTestRoute(route *state) *routeHandler {
	rh := newRouteHandler(route, nil)
	ctx, _ := NewContext()

	rh.exec(ctx, nil)
	return rh
}

func TestDefineUnconditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 3, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return true }).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))

	rf := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return false }).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh = execTestRoute(rf.getRouteStateMachine())

	assert.Equal(t, 2, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineNestedConditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return true }).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return true }).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 6, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return true }).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		otherwise().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return false }).
		addNextStep(doActionTest, undoActionTest).
		otherwise().
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 5, rh.statemachine.context.GetVariable("HK"))
}

func TestDefineRoute(t *testing.T) {
	r := newRoute().
		addNextStep(doActionTest, undoActionTest).
		when(func(ctx context) bool { return true }).
		addNextStep(doActionTest, undoActionTest).
		addNextStep(doActionTest, undoActionTest).
		otherwise().
		addNextStep(doActionTest, undoActionTest).
		end().
		addNextStep(doActionTest, undoActionTest)

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 4, rh.statemachine.context.GetVariable("HK"))
}
