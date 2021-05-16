package orchestrator

type (
	Route interface {
		GetRouteId() string
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
