package nrpc

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/ds248a/nrpc/log"
	"github.com/ds248a/nrpc/util"
)

var DefaultHandler Handler = NewHandler()

// Defines message handler of middleware and method/router.
type HandlerFunc func(*Context)

// Saves all middleware and method/router handler funcs for every method by register order,
// all the funcs will be called one by one for every message.
type routerHandler struct {
	async    bool
	handlers []HandlerFunc
}

// Defines net message handler interface.
type Handler interface {
	Clone() Handler
	LogTag() string
	SetLogTag(tag string)

	// Registers handler which will be called when client connected.
	HandleConnected(onConnected func(*Client))
	// Will be called when client is connected.
	OnConnected(c *Client)

	// Registers handler which will be called when client is disconnected.
	HandleDisconnected(onDisConnected func(*Client))
	// Will be called when client is disconnected.
	OnDisconnected(c *Client)

	// Registers handler which will be called when client send queue is overstock.
	HandleOverstock(onOverstock func(c *Client, m *Message))
	// Will be called when client chSend is full.
	OnOverstock(c *Client, m *Message)

	// Registers handler which will be called when message dropped.
	HandleMessageDropped(onOverstock func(c *Client, m *Message))
	// Will be called when message is dropped.
	OnMessageDropped(c *Client, m *Message)

	// Registers handler which will be called when async message seq not found.
	HandleSessionMiss(onSessionMiss func(c *Client, m *Message))
	// Will be called when async message seq not found.
	OnSessionMiss(c *Client, m *Message)

	// Registers handler which will be called before Recv.
	BeforeRecv(h func(net.Conn) error)
	// Registers handler which will be called before Send.
	BeforeSend(h func(net.Conn) error)

	// Returns BatchRecv flag.
	BatchRecv() bool
	// Sets BatchRecv flag.
	SetBatchRecv(batch bool)
	// Returns BatchSend flag.
	BatchSend() bool
	// Sets BatchSend flag.
	SetBatchSend(batch bool)

	// Returns AsyncResponse flag.
	AsyncResponse() bool
	// Sets AsyncResponse flag.
	SetAsyncResponse(async bool)

	// Wraps net.Conn to Read data with io.Reader.
	WrapReader(conn net.Conn) io.Reader
	// Registers reader wrapper for net.Conn.
	SetReaderWrapper(wrapper func(conn net.Conn) io.Reader)

	// Reads a message from a client.
	Recv(c *Client) (*Message, error)
	// Writes buffer data to a connection.
	Send(c net.Conn, buffer []byte) (int, error)
	// Writes multiple buffer data to a connection.
	SendN(conn net.Conn, buffers net.Buffers) (int, error)

	// Returns client's read buffer size.
	RecvBufferSize() int
	// Sets client's read buffer size.
	SetRecvBufferSize(size int)

	// Returns client's send queue channel capacity.
	SendQueueSize() int
	// Sets client's send queue channel capacity.
	SetSendQueueSize(size int)

	// Registers method/router handler middleware.
	Use(h HandlerFunc)

	// UseCoder registers message coding middleware,
	// coder.Encode will be called before message send,
	// coder.Decode will be called after message recv.
	UseCoder(coder MessageCoder)

	// Coders returns coding middlewares.
	Coders() []MessageCoder

	// Handle registers method/router handler.
	//
	// If pass a Boolean value of "true", the handler will be called asynchronously in a new goroutine,
	// Else the handler will be called synchronously in the client's reading goroutine one by one.
	Handle(m string, h HandlerFunc, args ...interface{})

	// HandleNotFound registers "" method/router handler,
	// It will be called when mothod/router is not found.
	HandleNotFound(h HandlerFunc)

	// OnMessage finds method/router middlewares and handler, then call them one by one.
	OnMessage(c *Client, m *Message)

	// GetBuffer makes a buffer by size.
	GetBuffer(size int) []byte

	// SetBufferFactory registers buffer maker.
	SetBufferFactory(f func(int) []byte)
}

// ------------------
//   Default handler
// ------------------

type handler struct {
	logtag         string
	batchRecv      bool
	batchSend      bool
	asyncResponse  bool
	recvBufferSize int
	sendQueueSize  int

	onConnected      func(*Client)
	onDisConnected   func(*Client)
	onOverstock      func(c *Client, m *Message)
	onMessageDropped func(c *Client, m *Message)
	onSessionMiss    func(c *Client, m *Message)

	beforeRecv    func(net.Conn) error
	beforeSend    func(net.Conn) error
	bufferFactory func(int) []byte

	wrapReader func(conn net.Conn) io.Reader

	middles   []HandlerFunc
	msgCoders []MessageCoder

	routes map[string]*routerHandler
}

func (h *handler) Clone() Handler {
	cp := *h
	cp.middles = make([]HandlerFunc, len(h.middles))
	copy(cp.middles, h.middles)

	cp.msgCoders = make([]MessageCoder, len(h.msgCoders))
	copy(cp.msgCoders, h.msgCoders)

	cp.routes = map[string]*routerHandler{}
	for k, v := range h.routes {
		rh := &routerHandler{
			async:    v.async,
			handlers: make([]HandlerFunc, len(v.handlers)),
		}
		copy(rh.handlers, v.handlers)
		cp.routes[k] = rh
	}

	return &cp
}

func (h *handler) LogTag() string {
	return h.logtag
}

func (h *handler) SetLogTag(tag string) {
	h.logtag = tag
}

func (h *handler) HandleConnected(onConnected func(*Client)) {
	if onConnected == nil {
		return
	}
	pre := h.onConnected
	h.onConnected = func(c *Client) {
		if pre != nil {
			pre(c)
		}
		onConnected(c)
	}
}

func (h *handler) OnConnected(c *Client) {
	if h.onConnected != nil {
		h.onConnected(c)
	}
}

func (h *handler) HandleDisconnected(onDisConnected func(*Client)) {
	if onDisConnected == nil {
		return
	}
	pre := h.onDisConnected
	h.onDisConnected = func(c *Client) {
		if pre != nil {
			pre(c)
		}
		onDisConnected(c)
	}
}

func (h *handler) OnDisconnected(c *Client) {
	if h.onDisConnected != nil {
		h.onDisConnected(c)
	}
}

func (h *handler) HandleOverstock(onOverstock func(c *Client, m *Message)) {
	h.onOverstock = onOverstock
}

func (h *handler) OnOverstock(c *Client, m *Message) {
	if h.onOverstock != nil {
		h.onOverstock(c, m)
	}
}

func (h *handler) HandleMessageDropped(onMessageDropped func(c *Client, m *Message)) {
	h.onMessageDropped = onMessageDropped
}

func (h *handler) OnMessageDropped(c *Client, m *Message) {
	if h.onMessageDropped != nil {
		h.onMessageDropped(c, m)
	}
}

func (h *handler) HandleSessionMiss(onSessionMiss func(c *Client, m *Message)) {
	h.onSessionMiss = onSessionMiss
}

func (h *handler) OnSessionMiss(c *Client, m *Message) {
	if h.onSessionMiss != nil {
		h.onSessionMiss(c, m)
	}
}

func (h *handler) BeforeRecv(hb func(net.Conn) error) {
	h.beforeRecv = hb
}

func (h *handler) BeforeSend(hs func(net.Conn) error) {
	h.beforeSend = hs
}

func (h *handler) BatchRecv() bool {
	return h.batchRecv
}

func (h *handler) SetBatchRecv(batch bool) {
	h.batchRecv = batch
}

func (h *handler) BatchSend() bool {
	return h.batchSend
}

func (h *handler) SetBatchSend(batch bool) {
	h.batchSend = batch
}

func (h *handler) AsyncResponse() bool {
	return h.asyncResponse
}

func (h *handler) SetAsyncResponse(async bool) {
	h.asyncResponse = async
}

func (h *handler) WrapReader(conn net.Conn) io.Reader {
	if h.wrapReader != nil {
		return h.wrapReader(conn)
	}
	return conn
}

func (h *handler) SetReaderWrapper(wrapper func(conn net.Conn) io.Reader) {
	h.wrapReader = wrapper
}

func (h *handler) RecvBufferSize() int {
	return h.recvBufferSize
}

func (h *handler) SetRecvBufferSize(size int) {
	h.recvBufferSize = size
}

func (h *handler) SendQueueSize() int {
	return h.sendQueueSize
}

func (h *handler) SetSendQueueSize(size int) {
	h.sendQueueSize = size
}

func (h *handler) Use(cb HandlerFunc) {
	if cb == nil {
		return
	}
	cbWithNext := func(ctx *Context) {
		cb(ctx)
		ctx.Next()
	}
	h.middles = append(h.middles, cbWithNext)
	for k, v := range h.routes {
		rh := &routerHandler{
			async:    v.async,
			handlers: make([]HandlerFunc, len(v.handlers)+1),
		}
		copy(rh.handlers, v.handlers)
		rh.handlers[len(v.handlers)] = cbWithNext
		h.routes[k] = rh
	}
}

func (h *handler) UseCoder(coder MessageCoder) {
	if coder != nil {
		h.msgCoders = append(h.msgCoders, coder)
	}
}

func (h *handler) Coders() []MessageCoder {
	return h.msgCoders
}

func (h *handler) Handle(method string, cb HandlerFunc, args ...interface{}) {
	if method == "" {
		panic(fmt.Errorf("empty('') method is reserved for [method not found], should use HandleNotFound to register '' handler"))
	}
	h.handle(method, cb, args...)
}

func (h *handler) HandleNotFound(cb HandlerFunc) {
	h.handle("", cb)
}

func (h *handler) handle(method string, cb HandlerFunc, args ...interface{}) {
	if h.routes == nil {
		h.routes = map[string]*routerHandler{}
	}
	if len(method) > MaxMethodLen {
		panic(fmt.Errorf("invalid method length %v(> MaxMethodLen %v)", len(method), MaxMethodLen))
	}

	if _, ok := h.routes[""]; !ok {
		rh := &routerHandler{
			async:    false,
			handlers: make([]HandlerFunc, len(h.middles)+1),
		}
		copy(rh.handlers, h.middles)
		rh.handlers[len(h.middles)] = func(ctx *Context) {
			ctx.Error(ErrMethodNotFound)
			ctx.Next()
		}
		h.routes[""] = rh
	}

	if _, ok := h.routes[method]; ok && method != "" {
		panic(fmt.Errorf("handler exist for method %v ", method))
	}

	async := h.AsyncResponse()
	if len(args) > 0 {
		if bv, ok := args[0].(bool); ok {
			async = bv
		}
	}
	rh := &routerHandler{
		async:    async,
		handlers: make([]HandlerFunc, len(h.middles)+1),
	}
	copy(rh.handlers, h.middles)
	rh.handlers[len(h.middles)] = func(ctx *Context) {
		cb(ctx)
		ctx.Next()
	}
	h.routes[method] = rh
}

func (h *handler) Recv(c *Client) (*Message, error) {
	var (
		err     error
		message *Message
	)

	if h.beforeRecv != nil {
		if err = h.beforeRecv(c.Conn); err != nil {
			return nil, err
		}
	}

	_, err = io.ReadFull(c.Reader, c.Head[:])
	if err != nil {
		return nil, err
	}

	message, err = c.Head.message(h)
	if err != nil {
		return nil, err
	}

	if message.Len() > HeadLen {
		_, err = io.ReadFull(c.Reader, message.Buffer[HeaderIndexBodyLenEnd:])
	}

	return message, err
}

func (h *handler) Send(conn net.Conn, buffer []byte) (int, error) {
	if h.beforeSend != nil {
		if err := h.beforeSend(conn); err != nil {
			return -1, err
		}
	}

	return conn.Write(buffer)
}

func (h *handler) SendN(conn net.Conn, buffers net.Buffers) (int, error) {
	if h.beforeSend != nil {
		if err := h.beforeSend(conn); err != nil {
			return -1, err
		}
	}

	n64, err := buffers.WriteTo(conn)
	return int(n64), err
}

func (h *handler) OnMessage(c *Client, msg *Message) {
	defer util.Recover()

	for i := len(h.msgCoders) - 1; i >= 0; i-- {
		msg = h.msgCoders[i].Decode(c, msg)
	}

	ml := msg.MethodLen()
	if ml <= 0 || ml > MaxMethodLen || ml > (msg.Len()-HeadLen) {
		log.Warn("%v OnMessage: invalid request method length %v, dropped", h.LogTag(), ml)
		return
	}

	cmd := msg.Cmd()
	switch cmd {
	case CmdRequest, CmdNotify:
		method := msg.method()
		if rh, ok := h.routes[method]; ok {
			ctx := newContext(c, msg, rh.handlers)
			if !rh.async {
				ctx.Next()
			} else {
				go ctx.Next()
			}
		} else {
			if cmd == CmdRequest {
				if rh, ok = h.routes[""]; ok {
					ctx := newContext(c, msg, rh.handlers)
					ctx.Next()
				} else {
					ctx := newContext(c, msg, rh.handlers)
					ctx.Error(ErrMethodNotFound)
				}
			}
			log.Warn("%v OnMessage: invalid method: [%v], no handler", h.LogTag(), method)
		}
		break
	case CmdResponse:
		if !msg.IsAsync() {
			seq := msg.Seq()
			session, ok := c.getSession(seq)
			if ok {
				session.done <- msg
			} else {
				h.OnSessionMiss(c, msg)
				log.Warn("%v OnMessage: session not exist or expired", h.LogTag())
			}
		} else {
			handler, ok := c.getAndDeleteAsyncHandler(msg.Seq())
			if ok {
				ctx := newContext(c, msg, nil)
				handler(ctx)
			} else {
				h.OnSessionMiss(c, msg)
				log.Warn("%v OnMessage: async handler not exist or expired", h.LogTag())
			}
		}
		break
	default:
		log.Warn("%v OnMessage: invalid cmd [%v]", h.LogTag(), msg.Cmd())
		break
	}
}

func (h *handler) GetBuffer(size int) []byte {
	if h.bufferFactory != nil {
		return h.bufferFactory(size)
	}
	return make([]byte, size)
}

func (h *handler) SetBufferFactory(f func(int) []byte) {
	h.bufferFactory = f
}

// Returns a default Handler implementation.
func NewHandler() Handler {
	h := &handler{
		logtag:         "[NRPC CLI]",
		batchRecv:      true,
		batchSend:      true,
		asyncResponse:  false,
		recvBufferSize: 8192,
		sendQueueSize:  4096,
	}
	h.wrapReader = func(conn net.Conn) io.Reader {
		return bufio.NewReaderSize(conn, h.recvBufferSize)
	}
	return h
}

// Sets default Handler.
func SetHandler(h Handler) {
	DefaultHandler = h
}

// Sets DefaultHandler's log tag.
func SetLogTag(tag string) {
	DefaultHandler.SetLogTag(tag)
}

// Registers default handler which will be called when client connected.
func HandleConnected(onConnected func(*Client)) {
	DefaultHandler.HandleConnected(onConnected)
}

// Registers default handler which will be called when client disconnected.
func HandleDisconnected(onDisConnected func(*Client)) {
	DefaultHandler.HandleDisconnected(onDisConnected)
}

// Registers default handler which will be called when client send queue is overstock.
func HandleOverstock(onOverstock func(c *Client, m *Message)) {
	DefaultHandler.HandleOverstock(onOverstock)
}

// Registers default handler which will be called when message dropped.
func HandleMessageDropped(onOverstock func(c *Client, m *Message)) {
	DefaultHandler.HandleMessageDropped(onOverstock)
}

// Registers default handler which will be called when async message seq not found.
func HandleSessionMiss(onSessionMiss func(c *Client, m *Message)) {
	DefaultHandler.HandleSessionMiss(onSessionMiss)
}

// Registers default handler which will be called before Recv.
func BeforeRecv(h func(net.Conn) error) {
	DefaultHandler.BeforeRecv(h)
}

// Registers default handler which will be called before Send.
func BeforeSend(h func(net.Conn) error) {
	DefaultHandler.BeforeSend(h)
}

// Returns default BatchRecv flag.
func BatchRecv() bool {
	return DefaultHandler.BatchRecv()
}

// Sets default BatchRecv flag.
func SetBatchRecv(batch bool) {
	DefaultHandler.SetBatchRecv(batch)
}

// Returns default BatchSend flag.
func BatchSend() bool {
	return DefaultHandler.BatchSend()
}

// Sets default BatchSend flag.
func SetBatchSend(batch bool) {
	DefaultHandler.SetBatchSend(batch)
}

// Returns default AsyncResponse flag.
func AsyncResponse() bool {
	return DefaultHandler.AsyncResponse()
}

// Sets default AsyncResponse flag.
func SetAsyncResponse(async bool) {
	DefaultHandler.SetAsyncResponse(async)
}

// Registers default reader wrapper for net.Conn.
func SetReaderWrapper(wrapper func(conn net.Conn) io.Reader) {
	DefaultHandler.SetReaderWrapper(wrapper)
}

// Returns default client's read buffer size.
func RecvBufferSize() int {
	return DefaultHandler.RecvBufferSize()
}

// Sets default client's read buffer size.
func SetRecvBufferSize(size int) {
	DefaultHandler.SetRecvBufferSize(size)
}

// Returns default client's send queue channel capacity.
func SendQueueSize() int {
	return DefaultHandler.SendQueueSize()
}

// Sets default client's send queue channel capacity.
func SetSendQueueSize(size int) {
	DefaultHandler.SetSendQueueSize(size)
}

// Default method/router handler middleware.
func Use(h HandlerFunc) {
	DefaultHandler.Use(h)
}

// UseCoder registers default message coding middleware,
// coder.Encode will be called before message send,
// coder.Decode will be called after message recv.
func UseCoder(coder MessageCoder) {
	DefaultHandler.UseCoder(coder)
}

// Handle registers default method/router handler.
//
// If pass a Boolean value of "true", the handler will be called asynchronously in a new goroutine,
// Else the handler will be called synchronously in the client's reading goroutine one by one.
func Handle(m string, h HandlerFunc, args ...interface{}) {
	DefaultHandler.Handle(m, h, args...)
}

// HandleNotFound registers default "" method/router handler,
// It will be called when mothod/router is not found.
func HandleNotFound(h HandlerFunc) {
	DefaultHandler.HandleNotFound(h)
}

// SetBufferFactory registers default buffer maker.
func SetBufferFactory(f func(int) []byte) {
	DefaultHandler.SetBufferFactory(f)
}
