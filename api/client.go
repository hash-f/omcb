package main

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

// client represents a connected client.
// Events are sent on the events channel and if the client
// cannot keep up with the messages, closeSlow is called.
type client struct {
	// events sent on this channel are sent to the websocket.
	events chan []byte

	// sendEvents controls whether the clients wants to recieve events or
	// just send events.
	// Used for creating bots that can send a lot of events to mimic traffic
	sendEvents bool

	// closeSlow closes the websocket if it is too slow to keep up
	closeSlow func()
}

func (c *client) toggleForwarding() {
	c.sendEvents = !c.sendEvents
}

func (c *client) startReadLoop(es *eventServer, conn *websocket.Conn, ctx context.Context) {
	// Goroutine to read messages from the WebSocket.
	for {
		typ, msg, err := conn.Read(ctx)

		if err != nil {
			es.deleteClient(c)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			es.logf("[%v] read error: %v", time.Now(), err)
			es.deleteClient(c)
			return
		}
		if typ == websocket.MessageBinary {
			es.handleCommand(msg, c)
		}
	}

}

func (c *client) startWriteLoop(conn *websocket.Conn, ctx context.Context) error {
	for {
		select {
		case msg := <-c.events:
			err := writeTimeout(ctx, time.Second*5, conn, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
