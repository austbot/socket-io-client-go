package socketio_client

import (
	"encoding/json"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/austbot/socket-io-client-go/transport"
	"github.com/cenkalti/backoff"
)

type Client struct {
	NameSpace *string
	url       string
	transport string
	stream    transport.Stream
	handlers  map[string]func(client *Client, data []string)
	mutex     sync.RWMutex
	isClosed  bool
}

func (c *Client) Connect(urlStr string, transport string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if len(u.Path) == 0 {
		u.Path = "socket.io"
	}
	c.url = u.String()
	c.isClosed = false
	c.transport = transport
	return c.doConnect(false)
}

func (c *Client) doConnect(reConnect bool) error {
	if reConnect {
		c.stream.Close()
		c.handleEvent("disconnect", []string{})
	}

	switch c.transport {
	case "polling":
		polling := transport.NewPolling(c.url, c.NameSpace)
		c.stream = &polling
	case "websocket":
		websocket := transport.NewWebsocket(c.url, c.NameSpace)
		c.stream = &websocket
	default:
		polling := transport.NewPolling(c.url, c.NameSpace)
		c.stream = &polling
	}

	if err := c.stream.Open(); err != nil {
		return err
	}
	go c.handleIncoming()
	return nil
}

func (c *Client) DisConnect() {
	c.isClosed = true
	c.stream.Close()
}

func (c *Client) Emit(name string, payload interface{}) error {
	msg, err := encode(name, c.NameSpace, payload)
	if err != nil {
		return err
	}
	return c.stream.Write(*msg)
}

func (c *Client) On(name string, handler func(client *Client, data []string)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.handlers == nil {
		c.handlers = make(map[string]func(client *Client, data []string))
	}

	c.handlers[name] = handler
}

func encode(name string, nameSpace *string, payload interface{}) (*string, error) {
	message := "42"
	if nameSpace != nil {
		message += *nameSpace + ","
	}

	if payload != nil {
		p, err := json.Marshal(&payload)
		if err != nil {
			return nil, err
		}

		message += "[\"" + name + "\"," + string(p) + "]"
	} else {
		message += "[\"" + name + "\"]"
	}

	return &message, nil
}

func (c *Client) handleIncoming() {

	for {
		if c.isClosed {
			break
		}

		msgs, err := c.stream.Read()
		if err != nil {
			go c.reConnect()
			break
		}

		if len(msgs) == 0 {
			continue
		}

		c.handleEvent(msgs[0], msgs[1:])
	}

}

func (c *Client) handleEvent(name string, msgs []string) {
	c.mutex.RLock()
	hStar, okStar := c.handlers["*"]
	handler, ok := c.handlers[name]
	c.mutex.RUnlock()
	if okStar {
		hStar(c, msgs)
	}
	if ok {
		handler(c, msgs)
	}
}

func (c *Client) reConnect() {
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 10 * time.Minute

	err := backoff.Retry(
		func() error {
			return c.doConnect(true)
		},
		b)
	if err != nil {
		log.Fatalf("error after retrying: %v", err)
	}
}
