package orchestrator

type (
	Route interface {
		GetStartState() *state
		GetEndpoints() []*Endpoint
	}

	Endpoint struct {
		// route id
		To string

		// State endpoint
		State *state
	}
)
