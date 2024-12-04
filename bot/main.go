package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/coder/websocket"
)

func main() {
	var wg sync.WaitGroup
	counter := make(chan bool)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		i := 0
		for <-counter {
			i++
			if i%1000 == 0 {
				fmt.Printf("Sent message %d\n", i)
			}
		}
	}()

	for {
		for i := range 500 {
			fmt.Printf("Starting bot %d\n", i)
			wg.Add(1)
			go newWorker(&counter, &sigs, &wg)
			time.Sleep(10 * time.Millisecond)
		}
		select {
		case <-sigs:
			return
		default:
			wg.Wait()
		}
	}
}

func newWorker(counter *chan bool, sigs *chan os.Signal, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws://localhost:8000/subscribe", nil)
	if err != nil {
		fmt.Println("Error dailing ws", err)
	}
	defer c.CloseNow()
	toggleRead(c, ctx)
	// toggleRead(c, ctx)
	go read(c, ctx)

	for {
		select {
		case sig := <-*sigs:
			fmt.Println("Got signal ", sig)
			c.Close(websocket.StatusNormalClosure, "")
			*sigs <- os.Interrupt
			return
		default:
			writer, err := c.Writer(ctx, websocket.MessageBinary)
			if err != nil {
				fmt.Println("error writing to ws ", err)
				return
			}
			action := rand.IntN(100) % 2
			// action := 0
			id := rand.IntN(10_000)

			msg := make([]byte, 5)
			msg[0] = byte(action)
			msg[1] = byte(id >> 24)
			msg[2] = byte(id >> 16)
			msg[3] = byte(id >> 8)
			msg[4] = byte(id)
			// fmt.Println("S", msg)
			err = binary.Write(writer, binary.BigEndian, msg)
			if err != nil {
				fmt.Println("Error writing message", err)
			}
			writer.Close()
			*counter <- true
			time.Sleep(1000 * time.Millisecond)
		}
	}
}

func read(c *websocket.Conn, ctx context.Context) {

	for {
		_, msg, err := c.Read(ctx)

		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			fmt.Printf("[%v] read error: %v\n", time.Now(), err)
			fmt.Println("read error: ", websocket.CloseStatus(err))

			return
		}

		fmt.Println("R", msg)
	}
}

func toggleRead(c *websocket.Conn, ctx context.Context) {
	writer, err := c.Writer(ctx, websocket.MessageBinary)
	if err != nil {
		fmt.Println("error toggling read ", err)
		return
	}
	action := 3

	msg := make([]byte, 1)
	msg[0] = byte(action)
	err = binary.Write(writer, binary.BigEndian, msg)
	if err != nil {
		fmt.Println("Error writing message", err)
	}
	writer.Close()
}
