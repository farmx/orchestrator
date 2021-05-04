package orchestrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type passStep struct {
	TransactionalStep
}

func (ps *passStep) DoAction(ctx *context) error {
	fv := ctx.GetVariable("A")

	if fv == nil {
		ctx.SetVariable("A", 1)
		return nil
	}

	ctx.SetVariable("A", fv.(int) + 1)
	return nil
}

func (ps *passStep) UndoAction(ctx context) {
	ctx.SetVariable("A", ctx.GetVariable("A").(int) - 1)
}

type bPassStep struct {
	TransactionalStep
}

func (ps *bPassStep) DoAction(ctx *context) error {
	fv := ctx.GetVariable("B")

	if fv == nil {
		ctx.SetVariable("B", 1)
		return nil
	}

	ctx.SetVariable("B", fv.(int) + 1)
	return nil
}

func (ps *bPassStep) UndoAction(ctx context) {
	ctx.SetVariable("A", ctx.GetVariable("B").(int) - 1)
}

func TestOrchestrator_Exec(t *testing.T) {
	aRoute := "A_ROUTE"
	bRoute := "B_ROUTE"

	orch := NewOrchestrator()
	orch.
		From(aRoute).
		AddStep(&passStep{}).
		When(func(ctx context) bool {
			return true
		}, &passStep{}).To(bRoute).
		End(&passStep{})

	orch.From(bRoute).AddStep(&bPassStep{})

	ctx, _ := NewContext()
	orch.Exec(aRoute, ctx, nil)

	assert.Equal(t, 2, ctx.GetVariable("A"))
	assert.Equal(t, 1, ctx.GetVariable("B"))
}
