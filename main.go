// websocket client.go
package main

import (
	"encoding/json"
	"fmt"
	"golang-ws-client/util"
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
var sendChan chan map[string]interface{}

/*
type rcvMessage struct {
	TableID     int    `json:"tableID"`
	UserID      string `json:"userID"`
	Status      string `json:"status"` //system auto flag
	ConnType    string `json:"connType"`
	IsActivated bool   `json:"isActivated"`
	Round       int    `json:"round"`
	SeatID      int    `json:"seatID"`
	Betvol      int    `json:"betvol"`
	Greeting    string `json:"greeting"`
} */

type RoomMsg struct {
	TID      int       `json:"tID"`
	Name     string    `json:"name"`
	MsgType  string    `json:"msgType"`
	Type     string    `json:"type"`
	SeatID   int       `json:"seatID"`
	Bvol     int       `json:"bvol"`
	Balance  int       `json:"balance"`
	FID      int       `json:"fID"`
	Status   [9]string `json:"status"`
	Types    [9]string `json:"tpyes"`
	Names    [9]string `json:"names"`
	Balances [9]int    `json:"balances"`
}

type Cards struct {
	CardsPoints [27]int   `json:"cardsPoints"`
	CardsSuits  [27]int   `json:"cardsSuits"`
	Cardstypes  [9]string `json:"cardstypes"`
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
} */

func main() {

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	sendChan = make(chan map[string]interface{})

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
		// case <-time.After(time.Duration(1) * time.Second * 60):
		case sendMsg := <-sendChan:

			// Send an next player packet if needed
			err := conn.WriteJSON(sendMsg)
			if err != nil {
				log.Println("Error during writing the json data to websocket:", err)
				return
			}
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Test message from Golang ws client every 60 secs"))
			// if err != nil {
			//	log.Println("Error during writing to websocket:", err)
			//	return
			// }

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

/*
// connType: NONE,JOINED,WAITING,ACTIVATE,BNEXT,TIMEOUT,CLOSE
// Create Next player info for sending
func createNextPlayerMsg(roomMsg [9]RoomMsg, seatID int) RoomMsg {
	var seatIDNext int
	var nextPlayerMsg RoomMsg

	i := seatID

	for {
		i++
		if i == TABLE_PLAYERS_MAX {
			i = 0
		}

		if roomMsg[i].ConnType == "" || roomMsg[i].ConnType == "NONE" {
			continue
		} else if i == seatID || roomMsg[i].ConnType != "" {
			seatIDNext = i
			break
		}
	}

	// fmt.Println("seatIDNext", seatIDNext)
	nextPlayerMsg = roomMsg[seatIDNext]
	nextPlayerMsg.IsActivated = true
	if nextPlayerMsg.Status == "AUTO" {
		nextPlayerMsg.ConnType = "BNEXT"
		nextPlayerMsg.Greeting = "Hello"
		nextPlayerMsg.IsActivated = false
	}

	return nextPlayerMsg
}
*/

func addCardsInfo(cards Cards) Cards {
	players := util.GetPlayersCards(50000012, 9)
	// fmt.Println(players)

	for i := 0; i < 9; i++ {
		cards.Cardstypes[i] = players[i].Cardstype
		for j := 0; j < 3; j++ {
			cards.CardsPoints[3*i+j] = players[i].Cards[j].Points
			cards.CardsSuits[3*i+j] = players[i].Cards[j].Suits
		}
	}

	return cards
}

func tableInfoDevlivery(delay time.Duration, ch chan map[string]interface{}) {
	var t [VOL_TABLE_MAX]*time.Timer
	// delayAuto := 6 * time.Second
	// var TableUsers [VOL_TABLE_MAX][TABLE_PLAYERS_MAX]rcvMessage
	// var nextPlayerMsg rcvMessage
	var roomNextMsg RoomMsg
	var cards Cards

	sendMap := make(map[string]interface{})

	roomNextMsg.TID = 0
	roomNextMsg.Name = "UNKNOWN"
	roomNextMsg.MsgType = "NEW"
	roomNextMsg.Type = "NONE"
	roomNextMsg.SeatID = 0
	roomNextMsg.Bvol = 0
	roomNextMsg.Balance = 0
	roomNextMsg.FID = 0
	for i := 0; i < 9; i++ {
		roomNextMsg.Status[i] = "MANUAL"
		roomNextMsg.Types[i] = "NONE"
		roomNextMsg.Names[i] = "UNKNOWN"
		roomNextMsg.Balances[i] = 0
	}

	testIndex := 0

	for i := 0; i < 27; i++ {
		cards.CardsPoints[i] = i
		cards.CardsSuits[i] = i
	}

	// TablesUsersMaps := make([]map[string]int, VOL_TABLE_MAX)
	for i := 0; i < VOL_TABLE_MAX; i++ {
		// TablesUsersMaps[i] = make(map[string]int, TABLE_PLAYERS_MAX)
		t[i] = time.NewTimer(delay)
	}

	for {
		select {
		case rcv := <-ch:
			testIndex++
			cards = addCardsInfo(cards)
			log.Println("S:", rcv)
			_, isOk := rcv["msgType"]
			if isOk {
				arr, err := json.Marshal(rcv)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = json.Unmarshal(arr, &roomNextMsg)
				if err != nil {
					fmt.Println(err)
					return
				}

			} else {
				log.Println("msgType not found", roomNextMsg)
			}

			// clear map
			for k := range rcv {
				delete(rcv, k)
			}
			t[0].Reset(delay)
			/*nextPlayerMsg = rcv

			nextPlayerMsg.SeatID = 1
			nextPlayerMsg.Status[nextPlayerMsg.SeatID] = "AUTO"
			t[rcv.TID].Reset(delay)
			*/

			/*
				// save the users info of the specific table
				TableUsers[rcv.TableID][rcv.SeatID] = rcv

				fmt.Println(rcv, "--Received Msg")
				// fmt.Println(TableUsers)
				nextPlayerMsg = createNextPlayerMsg(TableUsers[rcv.TableID], rcv.SeatID)

				if nextPlayerMsg.Status == "AUTO" {
					t[rcv.TableID].Reset(delayAuto)
					// fmt.Println("Set dealy for ATUO", delayAuto)
				} else {
					t[rcv.TableID].Reset(delay)
				} */
			continue
		case <-t[0].C:

			// delete map before assigned
			for k := range sendMap {
				delete(sendMap, k)
			}

			log.Println("textIndex", testIndex)
			if testIndex%2 == 0 {
				tmprec, _ := json.Marshal(&cards)
				json.Unmarshal(tmprec, &sendMap)

			} else {
				tmprec, _ := json.Marshal(&roomNextMsg)
				json.Unmarshal(tmprec, &sendMap)
			}

			// log.Println(nextPlayerMsg.Status[nextPlayerMsg.SeatID])
			/*
				if nextPlayerMsg.Status[nextPlayerMsg.SeatID] == "AUTO" {
					sendChan <- nextPlayerMsg
				} */
			log.Println("T0S:", sendMap)
			sendChan <- sendMap
			t[0].Reset(delay)
		case <-t[1].C:
			fmt.Println("T2 no new player message, repeat time interval:", delay)
			// t[1].Reset(delay)
		case <-t[2].C:
			fmt.Println("T3 no new player message, repeat time interval:", delay)
			// t[2].Reset(delay)
		}
	}
}

func receiveJsonHandler(connection *websocket.Conn) {
	var rcvMsg map[string]interface{}

	//var roomMsg RoomMsg
	ch := make(chan map[string]interface{})

	defer close(done)

	delay := 12 * time.Second
	go tableInfoDevlivery(delay, ch)

	for {
		err := connection.ReadJSON(&rcvMsg)
		if err != nil {
			log.Println("Received not JSON data!")
			continue
		}
		log.Println("R:", rcvMsg)
		ch <- rcvMsg
	}
}
