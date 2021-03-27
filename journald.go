package orchestrator

type journald struct {
}

// singleton
func getJournaldInstance() journald {
	return journald{}
}

func (j *journald) append(key string, data ...interface{}) {
}

func (j *journald) getLastEvent(key string) ([]interface{}, error) {
	return nil, nil
}
