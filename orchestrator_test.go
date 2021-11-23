package orchestrator

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestOrchestrator_Exec_HandoverBetweenRoutes(t *testing.T) {
	aRoute := "A_ROUTE"
	bRoute := "B_ROUTE"

	daa := func(ctx *context) error {
		fv := ctx.GetVariable("A")

		if fv == nil {
			ctx.SetVariable("A", 1)
			return nil
		}

		ctx.SetVariable("A", fv.(int)+1)
		return nil
	}

	uaa := func(ctx context) error {
		ctx.SetVariable("A", ctx.GetVariable("A").(int)-1)
		return nil
	}

	dab := func(ctx *context) error {
		fv := ctx.GetVariable("B")

		if fv == nil {
			ctx.SetVariable("B", 1)
			return nil
		}

		ctx.SetVariable("B", fv.(int)+1)
		return nil
	}

	uab := func(ctx context) error {
		ctx.SetVariable("A", ctx.GetVariable("B").(int)-1)
		return nil
	}

	orch := NewOrchestrator()
	ar := NewTransactionalRoute(aRoute).
		AddNextStep("1", daa, uaa).
		When(func(ctx context) bool { return true }).
		AddNextStep("when_1", daa, uaa).To(bRoute).
		End().
		AddNextStep("2", daa, uaa)

	br := NewTransactionalRoute(bRoute).AddNextStep("1", dab, uab)

	ctx, _ := NewContext()
	_ = orch.Register(ar)
	_ = orch.Register(br)

	_ = orch.Initialization(nil)
	orch.Exec(aRoute, ctx, nil)

	assert.Equal(t, 2, ctx.GetVariable("A"))
	assert.Equal(t, 1, ctx.GetVariable("B"))
}

func TestOrchestrator_Exec_RollbackBetweenTransactionalRoute(t *testing.T) {
	aRoute := "A_ROUTE"
	bRoute := "B_ROUTE"

	daa := func(ctx *context) error {
		ctx.SetVariable("STATE", "DAA")
		return nil
	}

	uaa := func(ctx context) error {
		log.Print("UAA")
		return nil
	}

	dab := func(ctx *context) error {
		ctx.SetVariable("STATE", "UAB")
		return errors.New("fake error")
	}

	uab := func(ctx context) error {
		log.Print("UAB")
		return nil
	}

	errChan := make(chan error)
	go func() {
		for err := range errChan {
			assert.NotNil(t, err)
		}
	}()

	orch := NewOrchestrator()
	ar := NewTransactionalRoute(aRoute).AddNextStep("1", daa, uaa).To(bRoute)
	br := NewTransactionalRoute(bRoute).AddNextStep("1", dab, uab)

	ctx, _ := NewContext()
	_ = orch.Register(ar)
	_ = orch.Register(br)

	_ = orch.Initialization(nil)
	orch.Exec(aRoute, ctx, errChan)
}
