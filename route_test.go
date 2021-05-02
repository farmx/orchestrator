package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type alwaysPassTransactionMock struct {
	TransactionalStep
}

func (aptm *alwaysPassTransactionMock) DoAction(ctx *context) error {
	if ctx.getVariable("HK") == nil {
		ctx.setVariable("HK", 0)
	}

	ctx.setVariable("HK", ctx.getVariable("HK").(int) + 1)

	log.Print("do action")
	return nil
}

func (aptm *alwaysPassTransactionMock) UndoAction(ctx context) {

}

func TestDefineUnconditionalRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 3, r.statemachine.context.getVariable("HK"))
}

func TestDefineConditionalRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return true},
			&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 5, r.statemachine.context.getVariable("HK"))

	rf := NewRoute("sample_f").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return false},
			&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{})

	ctxf, _ := NewContext()
	rf.Exec(*ctxf)

	assert.Equal(t, 2, rf.statemachine.context.getVariable("HK"))
}

func TestDefineNestedConditionalRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return true},
			&alwaysPassTransactionMock{}).
			When(func(ctx context) bool {return true},
				&alwaysPassTransactionMock{}).
				AddNextStep(&alwaysPassTransactionMock{}).
				AddNextStep(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 6, r.statemachine.context.getVariable("HK"))
}

func TestDefineConditionWithOtherwiseRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return true},
			&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{}).
		Otherwise(&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 4, r.statemachine.context.getVariable("HK"))
}

func TestDefineConditionWithOtherwiseAndEndRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return false},
			&alwaysPassTransactionMock{}).
		Otherwise(&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 5, r.statemachine.context.getVariable("HK"))
}

func TestDefineRoute(t *testing.T) {
	r := NewRoute("sample").
		AddNextStep(&alwaysPassTransactionMock{}).
		When(func(ctx context) bool {return true},
			&alwaysPassTransactionMock{}).
			AddNextStep(&alwaysPassTransactionMock{}).
		Otherwise(&alwaysPassTransactionMock{}).
		End(&alwaysPassTransactionMock{})

	ctx, _ := NewContext()
	r.Exec(*ctx)

	assert.Equal(t, 4, r.statemachine.context.getVariable("HK"))
}
