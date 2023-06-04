package transport

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fasthttp/websocket"
)

type Websocket struct {
	conn      *websocket.Conn
	url       string
	nameSpace *string
	sid       string
	isClosed  bool
	params    *TransportParams

	recv chan []string
	send chan []string
}

func NewWebsocket(url string, nameSpace *string) Websocket {
	url = strings.Replace(url, "http", "ws", 1)
	return Websocket{
		url:       url + "?EIO=4&transport=websocket",
		nameSpace: nameSpace,
		isClosed:  false,
		recv:      make(chan []string, ChannelBufferSize),
		send:      make(chan []string, ChannelBufferSize)}
}

func (w *Websocket) Open() error {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(w.url, http.Header{})
	if err != nil {
		return err
	}
	w.conn = conn

	msgType, reader, err := w.conn.NextReader()
	if err != nil {
		return err
	}
	if msgType != websocket.TextMessage {
		return errors.New("open packet should be text message")
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	body := string(data)

	if string(body[0]) != OpenPacket {
		return errors.New("first should be open packet")
	}

	// decode params
	var params TransportParams
	if err := json.Unmarshal([]byte(body[1:]), &params); err != nil {
		return err
	}
	w.params = &params

	// open namespace
	var openMsg string
	if w.nameSpace != nil {
		openMsg = OpenNamespace + *w.nameSpace + ","
	} else {
		openMsg = OpenNamespace
	}
	err = w.Write(openMsg)
	if err != nil {
		return err
	}

	w.recv <- []string{"connect"}

	return nil
}

func (w *Websocket) Read() (message []string, err error) {
	select {
	case message = <-w.recv:
		return message, nil
	default:
	}

	msgType, reader, err := w.conn.NextReader()
	if err != nil {
		return nil, err
	}
	if msgType != websocket.TextMessage {
		return nil, nil
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	payloadStr := string(data)

	if payloadStr == PingPacket {
		w.Write(PongPacket)
	}

	if strings.HasPrefix(payloadStr, NewMessage) {
		payloadStr = payloadStr[2:]
		if strings.Contains(payloadStr, *w.nameSpace) {
			payloadStr = strings.Split(payloadStr, *w.nameSpace+",")[1]
		}
		var message []string
		if err := json.Unmarshal([]byte(payloadStr), &message); err != nil {
			return nil, err
		}
		return message, nil
	}

	return nil, nil
}

func (w *Websocket) Write(message string) error {
	if w.conn == nil {
		return errors.New("websocket is not open")
	}

	writer, err := w.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}

	return writer.Close()
}
func (w *Websocket) Close() {

}
