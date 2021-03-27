package orchestrator

type journald interface {
	appendLog(data ...interface{})
	getLastEvent() ([]interface{}, error)
	shutdown()
}

type fileJournald struct {
}

// singleton
func getFileJournaldInstance(id string) journald {
	return &fileJournald{}
}

func (fj *fileJournald) appendLog(data ...interface{}) {

}

func (fj *fileJournald) getLastEvent() ([]interface{}, error) {
	return nil, nil
}

func (fj fileJournald) shutdown() {

}
