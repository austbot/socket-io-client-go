# Go SocketIO Client

This client was forked from [socket-io-client-go](https://github.com/socket-iox/socket-io-client-go) and modified to work with FastHTTP Websocket library which is essentially an actice fork of gorilla.
It supports Socket.io v4 clients currently and handles the handshake over a specific transport and does not support transport upgrades.