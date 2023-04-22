package gradio

import (
	"crypto/tls"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"io"
)

var (
	ErrNoConnect = errors.New("Failed to connect")
)

type Status int

const (
	Disconnected Status = iota
	Connecting
	Connected
	Error
)

type APIConnection struct {
	Status Status
	Ws     *websocket.Conn
}

func NewAPIConnection() *APIConnection {
	conn := APIConnection{
		Status: Disconnected,
		Ws:     nil,
	}

	return &conn
}

func (conn *APIConnection) Connect(host string) error {
	conn.Status = Connecting

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c, resp, err := dialer.Dial(
		"ws://"+host+"/queue/join",
		nil,
	)

	if err != nil {
		logrus.Error(err)

		if resp != nil {
			buf := make([]byte, 4096)
			_, err2 := resp.Body.Read(buf)
			if err2 == nil || err2 == io.EOF {
				logrus.Error(resp.Status + string(buf))
			}
		}

		conn.Status = Error
		return ErrNoConnect
	}

	conn.Ws = c
	conn.Status = Connected
	// caller should defer Ws.close()

	return nil
}

func (conn *APIConnection) Disconnect() {
	if conn.Status == Connected {
		conn.Ws.Close()
	}
}
