package orchestrator

import (
	"errors"
	"github.com/google/uuid"
)

type context struct {
	gid    string
	header map[string]interface{}
	body   interface{}
}

func NewContext(body interface{}) (*context, error) {
	guid := uuid.New().String()

	return NewContextWithGuid(guid, body)
}

func NewContextWithGuid(gid string, body interface{}) (*context, error) {
	if len(gid) < 1 {
		return nil, errors.New("GID is empty")
	}

	return &context{
		gid:  gid,
		body: body,
	}, nil
}

func (ctx *context) getHeader(key string) interface{} {
	return ctx.header[key]
}

func (ctx *context) setHeader(key string, value interface{}) {
	if ctx.header == nil {
		ctx.header = make(map[string]interface{})
	}

	ctx.header[key] = value
}

func (ctx *context) getBody() interface{} {
	return ctx.body
}

func (ctx *context) getGid() string {
	return ctx.gid
}
