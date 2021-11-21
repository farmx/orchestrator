package orchestrator

type routeState string

const (
	Main routeState = "MAIN"
	When routeState = "WHEN"
	Else routeState = "ELSE"
	End  routeState = "END"

	Condition int = 2
	Default   int = 1
)

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
