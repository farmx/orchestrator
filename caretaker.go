package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type caretaker interface {
	persist(id string, memento string) error
	get(id string) (string, error)
	shutdown() error
}

type fileCaretaker struct {
	caretaker
	f *os.File
}

type logStr struct {
	Timestamp string `json:"timestamp"`
	Id        string `json:"id"`
	Data      string `json:"data"`
}

var basePath = "."

func NewFileCareTacker(id string) (*fileCaretaker, error) {
	fileAddress := fmt.Sprintf("%s/%s.log", basePath, id)
	af, err := os.OpenFile(fileAddress,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644)

	if err != nil {
		return nil, err
	}

	return &fileCaretaker{
		f: af,
	}, nil
}

func (c *fileCaretaker) persist(id string, memento string) error {
	log, err := json.Marshal(logStr{
		Timestamp: time.Now().Format(time.RFC3339),
		Id:        id,
		Data:      memento,
	})

	if err != nil {
		return err
	}

	_, wErr := c.f.Write(append(log, []byte("\n")...))
	return wErr
}

// TODO: improve performance issue
// open new file descriptor to journal start to find the latest event ID from file head
func (c *fileCaretaker) get(id string) (string, error) {
	f, oErr := os.Open(c.f.Name())
	if oErr != nil {
		return "", oErr
	}

	var lastLog logStr
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var log logStr
		err := json.Unmarshal(scanner.Bytes(), &log)
		if err != nil {
			return "", err
		}

		if log.Id == id {
			lastLog = log
		}
	}

	return lastLog.Data, nil
}

func (c *fileCaretaker) shutdown() error {
	c.f.Sync()
	return c.f.Close()
}
