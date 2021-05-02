package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type alwaysPassTransactionMock struct {
	TransactionalStep
}

func (aptm *alwaysPassTransactionMock) DoAction(ctx *context) error {
	if ctx.getVariable("HK") == nil {
		ctx.setVariable("HK", 0)
	}

	ctx.setVariable("HK", ctx.getVariable("HK").(int)+1)
	return nil
}

func (aptm *alwaysPassTransactionMock) UndoAction(ctx context) {

}

func execTestRoute(route *state) *routeHandler {
	rh := newRouteHandler("sample", route, nil)
	ctx, _ := NewContext()

	rh.exec(*ctx)
	return rh
}

func TestDefineUnconditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 3, rh.statemachine.context.getVariable("HK"))
}

func TestDefineConditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return true },
			&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 5, rh.statemachine.context.getVariable("HK"))

	rf := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return false },
			&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh = execTestRoute(rf.getRouteStateMachine())

	assert.Equal(t, 2, rh.statemachine.context.getVariable("HK"))
}

func TestDefineNestedConditionalRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return true },
			&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return true },
			&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 6, rh.statemachine.context.getVariable("HK"))
}

func TestDefineConditionWithOtherwiseRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return true },
			&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		otherwise(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 4, rh.statemachine.context.getVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return false },
			&alwaysPassTransactionMock{}).
		otherwise(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 5, rh.statemachine.context.getVariable("HK"))
}

func TestDefineRoute(t *testing.T) {
	r := newRoute().
		addNextStep(&alwaysPassTransactionMock{}).
		when(func(ctx context) bool { return true },
			&alwaysPassTransactionMock{}).
		addNextStep(&alwaysPassTransactionMock{}).
		otherwise(&alwaysPassTransactionMock{}).
		end(&alwaysPassTransactionMock{})

	rh := execTestRoute(r.getRouteStateMachine())

	assert.Equal(t, 4, rh.statemachine.context.getVariable("HK"))
}
