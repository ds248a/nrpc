package nrpc

// ------------------
//   Websocket
// ------------------

type WebsocketConn interface {
	HandleWebsocket(func())
}
