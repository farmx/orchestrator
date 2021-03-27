package orchestrator

type journald struct {

}

// singleton
func getJournaldInstance() journald {
	return journald{}
}

func (j *journald) journal(key string, data ...interface{}) error {
	return nil
}

func (j *journald) getLastEvent(key string) ([]interface{}, error) {
	return nil, nil
}
