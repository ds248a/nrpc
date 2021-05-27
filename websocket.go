package nrpc

type WebsocketConn interface {
	HandleWebsocket(func())
}
