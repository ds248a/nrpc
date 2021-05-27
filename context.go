package nrpc

import (
	"math"
	"time"
)

// ------------------
//   Context
// ------------------

type Context struct {
	Client  *Client
	Message *Message
	values  map[interface{}]interface{}

	index    int
	handlers []HandlerFunc
}

// Returns value for key.
func (ctx *Context) Get(key interface{}) (interface{}, bool) {
	if len(ctx.values) == 0 {
		return nil, false
	}
	value, ok := ctx.values[key]
	return value, ok
}

// Sets key-value pair.
func (ctx *Context) Set(key interface{}, value interface{}) {
	if key == nil || value == nil {
		return
	}
	if ctx.values == nil {
		ctx.values = map[interface{}]interface{}{}
	}
	ctx.values[key] = value
}

// Returns values.
func (ctx *Context) Values() map[interface{}]interface{} {
	return ctx.values
}

// Returns body.
func (ctx *Context) Body() []byte {
	return ctx.Message.Data()
}

// Bind parses the body data and stores the result in the value pointed to by 'v'.
func (ctx *Context) Bind(v interface{}) error {
	msg := ctx.Message
	if msg.IsError() {
		return msg.Error()
	}
	if v != nil {
		data := msg.Data()
		switch vt := v.(type) {
		case *[]byte:
			*vt = data
		case *string:
			*vt = string(data)
		// case *error:
		// 	*vt = errors.New(util.BytesToStr(data))
		default:
			return ctx.Client.Codec.Unmarshal(data, v)
		}
	}
	return nil
}

// Responses a Message to the Client.
func (ctx *Context) Write(v interface{}) error {
	return ctx.write(v, false, TimeForever)
}

// Responses a Message to the Client with timeout.
func (ctx *Context) WriteWithTimeout(v interface{}, timeout time.Duration) error {
	return ctx.write(v, false, timeout)
}

// Responses an error Message to the Client.
func (ctx *Context) Error(v interface{}) error {
	return ctx.write(v, true, TimeForever)
}

// Calls next middleware or method/router handler.
func (ctx *Context) Next() {
	index := int(ctx.index)
	if index < len(ctx.handlers) {
		ctx.index++
		ctx.handlers[index](ctx)
	}
}

// Stops the one-by-one-calling of middlewares and method/router handler.
func (ctx *Context) Abort() {
	ctx.index = int(math.MaxInt8)
}

// Implements stdlib's Context.
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

// Implements stdlib's Context.
func (ctx *Context) Done() <-chan struct{} {
	return nil
}

// Implements stdlib's Context.
func (ctx *Context) Err() error {
	return nil
}

// Returns the value associated with this context for key, implements stdlib's Context.
func (ctx *Context) Value(key interface{}) interface{} {
	value, _ := ctx.Get(key)
	return value
}

func (ctx *Context) write(v interface{}, isError bool, timeout time.Duration) error {
	cli := ctx.Client
	if cli.IsAsync {
		return ctx.writeDirectly(v, isError)
	}
	req := ctx.Message
	if req.Cmd() != CmdRequest {
		return ErrContextResponseToNotify
	}
	if _, ok := v.(error); ok {
		isError = true
	}
	rsp := newMessage(CmdResponse, req.method(), v, isError, req.IsAsync(), req.Seq(), cli.Handler, cli.Codec, ctx.values)
	return cli.PushMsg(rsp, timeout)
}

func (ctx *Context) writeDirectly(v interface{}, isError bool) error {
	cli := ctx.Client
	req := ctx.Message
	if req.Cmd() != CmdRequest {
		return ErrContextResponseToNotify
	}
	if _, ok := v.(error); ok {
		isError = true
	}
	rsp := newMessage(CmdResponse, req.method(), v, isError, req.IsAsync(), req.Seq(), cli.Handler, cli.Codec, ctx.values)
	_, err := cli.Conn.Write(rsp.Buffer)
	return err
}

func newContext(cli *Client, msg *Message, handlers []HandlerFunc) *Context {
	return &Context{Client: cli, Message: msg, values: msg.values, index: 0, handlers: handlers}
}
