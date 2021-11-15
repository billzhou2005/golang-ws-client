// websocket client.go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const VOL_TABLE_MAX int = 3
const TABLE_PLAYERS_MAX int = 9

var done chan interface{}
var interrupt chan os.Signal

type rcvMessage struct {
	TableID  int    `json:"tableID"`
	ConnType string `json:"connType"`
	Status   string `json:"status"`
	UserID   string `json:"userID"`
	SeatID   int    `json:"seatID"`
	Betvol   int    `json:"betvol"`
	Greeting string `json:"greeting"`
}

type Player struct {
	NickName string `json:"nickName"`
	Sts      string `json:"sts"`
	SID      int    `json:"sID"`
	Vol      int    `json:"vol"`
	Tol      int    `json:"Tol"`
}

type RoomMsg struct {
	TID     int       `json:"tID"`
	SType   string    `json:"sType"`
	RType   string    `json:"rType"`
	UsID    int       `json:"usID"`
	FID     int       `json:"fID"`
	Res     string    `json:"res"`
	Players [6]Player `json:"players"`
}
type Card struct {
	Points int `json:"points"`
	Suits  int `json:"suits"`
}

type PlayerWithCards struct {
	Cards [3]Card `json:"cards"`
}

type RoomMsgWithCards struct {
	TID              int                `json:"tID"`
	SType            string             `json:"sType"`
	RType            string             `json:"rType"`
	UsID             int                `json:"usID"`
	FID              int                `json:"fID"`
	Res              string             `json:"res"`
	PlayersWithCards [6]PlayerWithCards `json:"playersWithCards"`
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
		fmt.Println(msg)
	}
}
func receiveJsonHandler(connection *websocket.Conn) {
	// var rcvMsg map[string]interface{}
	var roomMsg RoomMsg

	ch := make(chan rcvMessage)

	defer close(done)

	delay := 12 * time.Second
	go tableInfoDevlivery(delay, ch)

	for {
		err := connection.ReadJSON(&roomMsg)
		if err != nil {
			log.Println("Received test message every 60s or not Json data")
			continue
		}
		log.Println(roomMsg)

		// log.Println(ei.N(rcv).M("connType").StringZ())
		// ch <- rcv
	}
}
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
	// go receiveHandler(conn)
	go receiveJsonHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(1) * time.Second * 60):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte("Test message from Golang ws client every 60 secs"))
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

func tableInfoDevlivery(delay time.Duration, ch chan rcvMessage) {
	var t [VOL_TABLE_MAX]*time.Timer

	TablesUsersMaps := make([]map[string]int, VOL_TABLE_MAX)
	for i := 0; i < VOL_TABLE_MAX; i++ {
		TablesUsersMaps[i] = make(map[string]int, TABLE_PLAYERS_MAX)
		t[i] = time.NewTimer(delay)
	}

	for {
		select {
		case rcv := <-ch:
			fmt.Println("TableID", rcv.TableID, "UserID", rcv.UserID, "connType", rcv.ConnType)
			if rcv.ConnType == "CLOSE" {
				delete(TablesUsersMaps[rcv.TableID], rcv.UserID)
			} else {
				TablesUsersMaps[rcv.TableID][rcv.UserID] = rcv.SeatID
			}
			log.Println(TablesUsersMaps)
			t[rcv.TableID].Reset(delay)
			continue
		case <-t[0].C:
			fmt.Println("T1 no new player message, repeat time interval:", delay)
			t[0].Reset(delay)
		case <-t[1].C:
			fmt.Println("T2 no new player message, repeat time interval:", delay)
			t[1].Reset(delay)
		case <-t[2].C:
			fmt.Println("T3 no new player message, repeat time interval:", delay)
			t[2].Reset(delay)
		}
	}
}
