package nrpc

// ------------------
//   RPC Session
// ------------------

type rpcSession struct {
	seq  uint64
	done chan *Message
}

func newSession(seq uint64) *rpcSession {
	return &rpcSession{seq: seq, done: make(chan *Message, 1)}
}
