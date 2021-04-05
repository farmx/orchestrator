package orchestrator

import (
	"os"
	"sync"
)

type caretaker interface {
	persist(id string, memento string)
	get(id string) string
	shutdown()
}

type fileCaretaker struct {
	file *os.File
	mx   sync.RWMutex
}

func (f *fileCaretaker) persist(id string, memento string) {
	panic("implement me")
}

func (f *fileCaretaker) get(id string) string {
	panic("implement me")
}

func (f *fileCaretaker) shutdown() {
	panic("implement me")
}
