package main

import (
	"fmt"
	"time"

	socketio "github.com/austbot/socket-io-client-go"
)

func main() {
	nameSpace := "/server/realtime"
	client := socketio.Client{NameSpace: &nameSpace}
	client.On("connect", func(client *socketio.Client, data []string) {
		fmt.Println("connected")
		client.Emit("register", map[string]string{"key": "server-8fcb4ae3042ddf364d2a9fa16597b92dc3c22aa7"})
	})

	client.On("disconnect", func(client *socketio.Client, data []string) {
		fmt.Println("disconnect")
	})

	client.On("test", func(client *socketio.Client, data []string) {
		fmt.Println(data)
	})

	client.On("update", func(client *socketio.Client, data []string) {
		fmt.Println("update", data)
	})

	if err := client.Connect("https://featureprobe.io/server/realtime", "websocket"); err != nil {
		fmt.Println(err)
		return
	}

	for {
		time.Sleep(5 * time.Second)
	}
}
