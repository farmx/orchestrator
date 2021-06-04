package orchestrator

type (
	Route interface {
		GetRouteId() string
		GetStartState() *State
		GetEndpoints() []*Endpoint
	}

	Endpoint struct {
		// route id
		To string

		// State endpoint
		State *State
	}
)
