package orchestrator

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeSuccessStep struct {
	TransactionStep
}

func (f *fakeSuccessStep) process(ctx *context) error {
	return nil
}

func (f *fakeSuccessStep) failed(ctx context) {

}

type fakeFailedStep struct {
	TransactionStep
}

func (f *fakeFailedStep) process(ctx *context) error {
	return errors.New("fake failed")
}

func (f *fakeFailedStep) failed(ctx context) {

}

func TestAddNextStep(t *testing.T) {
	r := newRoute("atomicRoute")

	r.addNextStep(&fakeSuccessStep{})
	r.addNextStep(&fakeSuccessStep{})
	r.addNextStep(&fakeSuccessStep{})
	r.addNextStep(&fakeSuccessStep{})
	r.addNextStep(&fakeSuccessStep{})

	assert.Equal(t, 5, len(r.steps))
}

func TestHasNext(t *testing.T) {
	scenarios := []struct {
		state        transactionState
		step         int
		numberOfStep int
		expected     bool
	}{
		{
			state:        Start,
			step:         0,
			numberOfStep: 1,
			expected:     true,
		},
		{
			state:        Closed,
			step:         1,
			numberOfStep: 1,
			expected:     false,
		},
		{
			state:        InProgress,
			step:         1,
			numberOfStep: 2,
			expected:     true,
		},
		{
			state:        Rollback,
			step:         1,
			numberOfStep: 2,
			expected:     true,
		},
	}
	for _, scenario := range scenarios {

		r := newRoute("testRoute")
		r.state = scenario.state
		r.currentStep = scenario.step
		for i := 0; i < scenario.numberOfStep; i++ {
			r.addNextStep(&fakeSuccessStep{})
		}

		assert.Equal(t, scenario.expected, r.hasNext())
	}
}

func TestExecNextStep_rollbackOnProcessFailed(t *testing.T) {
	r := newRoute("route_id")
	ctx, _ := NewContext(nil)
	r.init(*ctx)
	r.addNextStep(&fakeSuccessStep{})
	r.addNextStep(&fakeFailedStep{})

	r.execNextStep()
	err := r.execNextStep()

	assert.NotNil(t, err)
	assert.Equal(t, r.state, Rollback)
	assert.Equal(t, r.status, Unknown)
}

func TestCreateMementoAndRestore_RestoreNewRouteWithAnotherRouteMemento(t *testing.T) {
	routeId := "ROUTE_ID"
	cStep := 1
	state := InProgress
	status := Unknown
	gid := "11"

	r := newRoute(routeId)
	r.currentStep = cStep
	r.state = state
	r.status = status
	r.ctx = &context{
		gid: gid,
	}

	mem := r.createMemento()

	r2 := newRoute("route_2")
	err := r2.restore(mem)

	assert.Nil(t, err)
	assert.Equal(t, routeId, r.id)
	assert.Equal(t, cStep, r.currentStep)
	assert.Equal(t, state, r.state)
	assert.Equal(t, cStep, r.currentStep)
	assert.Equal(t, status, r.status)
	assert.Equal(t, gid, r.ctx.gid)
}

func TestUpdateState_updateRouteStateOnSuccessAndFailedOnly(t *testing.T) {
	scenarios := []struct {
		numberOfStep int
		currentStep  int
		expected     struct {
			status transactionStatus
			state  transactionState
		}
	}{
		{
			numberOfStep: 2,
			currentStep:  2,
			expected: struct {
				status transactionStatus
				state  transactionState
			}{
				status: Success,
				state:  Closed,
			},
		},
		{
			numberOfStep: 2,
			currentStep:  -1,
			expected: struct {
				status transactionStatus
				state  transactionState
			}{
				status: Fail,
				state:  Closed,
			},
		},
		{
			numberOfStep: 2,
			currentStep:  1,
			expected: struct {
				status transactionStatus
				state  transactionState
			}{
				status: Unknown,
				state:  InProgress,
			},
		},
	}

	for _, scenario := range scenarios {

		r := newRoute("testRoute")
		r.state = InProgress
		r.status = Unknown
		r.currentStep = scenario.currentStep
		for i := 0; i < scenario.numberOfStep; i++ {
			r.addNextStep(&fakeSuccessStep{})
		}

		r.updateState()

		assert.Equal(t, scenario.expected.status, r.status)
		assert.Equal(t, scenario.expected.state, r.state)
	}
}
