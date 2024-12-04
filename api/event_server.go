package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	key         = "state"
	counterKey  = "totalCount"
	channelName = "send-click-events"
)

// eventServer accepts events from clients and forwards batched events to subscribers..
type eventServer struct {
	// eventBufferLen controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	eventBufferLen int

	// logf controls where logs are sent.
	// Defaults to log.Printf.
	logf func(f string, v ...interface{})

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	// clientsMu controlls access to the subscribers map
	clientsMu sync.Mutex

	// clients stores a map of all connected clients
	clients map[*client]struct{}

	// cacheClient stores a pointer to the current CacheClient
	cacheClient *RedisClient

	// originPatterns controlls the origins that are allowed to connect
	originPatterns string
}

// newEventServer constructs a chatServer with the defaults.
func newEventServer() *eventServer {
	es := &eventServer{
		eventBufferLen: 16,
		logf:           log.Printf,
		clients:        make(map[*client]struct{}),
		cacheClient:    newRedisClient(),
		originPatterns: "*",
	}
	es.serveMux.HandleFunc("/subscribe", es.subscribeHandler)
	es.serveMux.HandleFunc("/state", es.getStateHandler)
	es.serveMux.HandleFunc("/stats", es.getStatsHandler)

	// Set the state key if it is not set
	es.cacheClient.init(key)

	// Start a go routine to forward events recieved from cache pub sub
	go es.forwardEvents()

	return es
}

func (es *eventServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	es.serveMux.ServeHTTP(w, r)
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (es *eventServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := es.subscribe(w, r)
	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		es.logf("%v", err)
		return
	}
}

// getStateHandler handles the requests to the GET /state route
// It returns the state of all the checkboxes stored in the cache
func (es *eventServer) getStateHandler(w http.ResponseWriter, r *http.Request) {
	// @todo: Limit allowed origins to specified origins
	w.Header().Set("Access-Control-Allow-Origin", es.originPatterns)
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	val, err := es.cacheClient.get(key)
	if err != nil {
		es.logf("%v", err)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprint(w, val)
}

// getStateHandler handles the requests to the GET /state route
// It returns the state of all the checkboxes stored in the cache
func (es *eventServer) getStatsHandler(w http.ResponseWriter, r *http.Request) {
	// @todo: Limit allowed origins to specified origins
	w.Header().Set("Access-Control-Allow-Origin", es.originPatterns)
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	val, err := es.cacheClient.get(counterKey)
	if err != nil {
		es.logf("%v", err)
	}
	count, err := strconv.Atoi(val)
	if err != nil {
		count = 0
	}

	w.Header().Set("Content-Type", "text/json")

	stats := map[string]int{"total": count}
	statsJson, err := json.Marshal(stats)
	resp := string(statsJson)
	if err != nil {
		resp = "{\"count\": 0}"
	}

	fmt.Fprint(w, resp)
}

// subscribe subscribes the given WebSocket to all events.
// It creates a subscriber with a buffered msgs chan to give some room to slower
// connections and then registers the subscriber. It then listens for all events
// and writes them to the WebSocket. If the context is cancelled or
// an error occurs, it returns and deletes the subscription.
func (es *eventServer) subscribe(w http.ResponseWriter, r *http.Request) error {
	var mu sync.Mutex
	var conn *websocket.Conn
	var closed bool
	c := &client{
		events:     make(chan []byte, es.eventBufferLen),
		sendEvents: true,
		closeSlow: func() {
			mu.Lock()
			defer mu.Unlock()
			closed = true
			if conn != nil {
				conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
			}
		},
	}
	es.addClient(c)
	defer es.deleteClient(c)

	temp, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{es.originPatterns},
	})
	if err != nil {
		return err
	}
	mu.Lock()
	if closed {
		mu.Unlock()
		return net.ErrClosed
	}
	conn = temp
	mu.Unlock()
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go c.startReadLoop(es, conn, ctx)

	err = c.startWriteLoop(conn, ctx)

	return err
}

func (es *eventServer) publishBinary(msg []byte) {
	// @todo: Check if this lock would affect incoming connections
	es.clientsMu.Lock()
	defer es.clientsMu.Unlock()
	// Loop through all clients and send the binary message
	for s := range es.clients {
		if !s.sendEvents {
			continue
		}
		select {
		case s.events <- msg:
		default:
			go s.closeSlow()
		}
	}
}

// addClient registers a client.
func (es *eventServer) addClient(c *client) {
	es.clientsMu.Lock()
	es.clients[c] = struct{}{}
	es.logf("Total %d subscribers", len(es.clients))
	es.clientsMu.Unlock()

}

// deleteClient deletes the given client.
func (es *eventServer) deleteClient(c *client) {
	es.clientsMu.Lock()
	delete(es.clients, c)
	es.clientsMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageBinary, msg)
}

func (es *eventServer) forwardEvents() {
	pubsub := es.cacheClient.subscribe(channelName)
	defer pubsub.Close()
	es.logf("Setting up event forwarding")
	ch := pubsub.Channel()
	bucket := make([]byte, 0)
	start := time.Now()
	time.Since(start)

	// @todo: Change the batch size based on traffic.
	// Store traffic per second in redis and set that as the batch size
	batchSize := 1
	bytesPerMsg := 5
	for msg := range ch {
		bucket = append(bucket, []byte(msg.Payload)...)
		if len(bucket) == batchSize*bytesPerMsg {
			es.publishBinary(bucket)
			bucket = nil

		}
	}
}

func (es *eventServer) handleCommand(msg []byte, c *client) {
	if len(msg) == 0 {
		es.logf("Empty command. Message length is 0")
		return
	}

	action := msg[0]
	switch action {
	case 0:
		// action: 0 means an uncheck event
		fallthrough
	case 1:
		// action: 1 means a check event
		if len(msg) == 5 {
			id := int32(binary.BigEndian.Uint32(msg[1:]))
			if id >= 1_000_000 {
				es.logf("Id: %d is not in range. Something Fishy?", id)
			} else if action != 0 && action != 1 {
				es.logf("Action: %d is not in range. Something Fishy?", action)
			} else {
				es.cacheClient.incr(counterKey)
				es.cacheClient.setBit(key, int64(id), int(action))
				if err := es.cacheClient.publish(channelName, string(msg)); err != nil {
					panic(err)
				}
			}
		}
	case 3:
		// action: 3 toggles event forwarding
		es.clientsMu.Lock()
		c.toggleForwarding()
		es.clientsMu.Unlock()
	}
}
