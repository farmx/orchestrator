package orchestrator

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type context struct {
	gid       string
	lock      sync.Mutex
	constant  map[string]interface{}
	variables map[string]interface{}
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
		constant:  make(map[string]interface{}),
		variables: make(map[string]interface{}),
	}, nil
}

func (ctx *context) getConstant(key string) interface{} {
	return ctx.constant[key]
}

func (ctx *context) setConstant(key string, value interface{}) error {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()
	if ctx.constant[key] != nil {
		return errors.New(fmt.Sprintf("key %s was reservide", key))
	}

	ctx.constant[key] = value
	return nil
}

func (ctx *context) setVariable(key string, value interface{}) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	ctx.variables[key] = value
}

func (ctx *context) getVariable(key string) interface{} {
	return ctx.variables[key]
}

func (ctx *context) getGid() string {
	return ctx.gid
}
