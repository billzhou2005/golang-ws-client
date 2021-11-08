// websocket client.go
package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const TABLEPRIMARY string = "100012"
const TABLEMIDDLE string = "200012"
const TABLEADVANCED string = "300012"

var done chan interface{}
var interrupt chan os.Signal
var rcvMessage struct {
	TableID  int    `json:"tableID"`
	UserID   string `json:"userID"`
	SeatID   int    `json:"seatID"`
	ConnType string `json:"connType"`
	Status   string `json:"status"`
	Betvol   int    `json:"betvol"`
	Greeting string `json:"greeting"`
}

func receiveJsonHandler(connection *websocket.Conn) {
	//var TablesUserMap []map[string]int
	//var TablesUserMap []map[string]int
	TablesUserMap := make([]map[string]int, 3)
	TablesUserMap[0] = make(map[string]int, 9)
	TablesUserMap[1] = make(map[string]int, 9)
	TablesUserMap[2] = make(map[string]int, 9)

	defer close(done)
	for {

		err := connection.ReadJSON(&rcvMessage)
		if err != nil {
			log.Println("Received not JSON data")
			continue
		}
		// monitor the number of players on a specific table
		if rcvMessage.TableID < 3 {
			if rcvMessage.ConnType == "CLOSE" {
				delete(TablesUserMap[rcvMessage.TableID], rcvMessage.UserID)
			} else {
				TablesUserMap[rcvMessage.TableID][rcvMessage.UserID] = rcvMessage.SeatID
			}
			log.Println(TablesUserMap)
		} else {
			log.Println("not TableID < 2, invalid TableID", rcvMessage.TableID)
		}

	}
}

/*
func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}
*/
func main() {

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := "ws://140.143.149.188:9080" + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveJsonHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 50000):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte("Test message from Golang ws client every 50 secs"))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}
