package orchestrator

import (
	"errors"
	"github.com/google/uuid"
	"sync"
)

const DefaultVersion = "v1"

type context struct {
	gid       string
	lock      sync.Mutex
	variables map[string]row
}

type row struct {
	version string
	value   interface{}
}

func NewContext() (*context, error) {
	guid := uuid.New().String()

	return NewContextWithGid(guid)
}

func NewContextWithGid(gid string) (*context, error) {
	if len(gid) < 1 {
		return nil, errors.New("GID is empty")
	}

	return &context{
		gid:       gid,
		lock:      sync.Mutex{},
		variables: make(map[string]row),
	}, nil
}

func (ctx *context) SetVariableWithVersion(key string, lastVersion string, newVersion string, value interface{}) error {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ver := ctx.variables[key].version
	if ver != "" && ver != lastVersion {
		return errors.New("invalid data version")
	}

	ctx.variables[key] = row{
		version: newVersion,
		value:   value,
	}

	return nil
}

func (ctx *context) SetVariable(key string, value interface{}) error {
	return ctx.SetVariableWithVersion(key, DefaultVersion, DefaultVersion, value)
}

func (ctx *context) GetVariable(key string) interface{} {
	return ctx.variables[key].value
}

func (ctx *context) GetGid() string {
	return ctx.gid
}
