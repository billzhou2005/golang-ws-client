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
var sendChan chan RoomMsg

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
	SeatID   int       `json:"seatID"`
	Bvol     int       `json:"bvol"`
	Balance  int       `json:"balance"`
	FID      int       `json:"fID"`
	Status   [9]string `json:"status"`
	Types    [9]string `json:"tpyes"`
	Names    [9]string `json:"names"`
	Balances [9]int    `json:"balances"`
	//	CardsPoints [27]int   `json:"cardsPoints"`
	//	CardsSuits  [27]int   `json:"cardsSuits"`
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
<<<<<<< HEAD
}
=======
} */

>>>>>>> d6972a01e97c63df6da4068c043b0596b05a2452
func main() {

	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully
	sendChan = make(chan RoomMsg)

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

/*
func addCardsInfo(roomMsg RoomMsg) RoomMsg {
	players := util.GetPlayersCards(50000012, 9)
	// fmt.Println(players)


	for i := 0; i < 9; i++ {
		for j := 0; j < 3; j++ {
			roomMsg.CardsPoints[3*i+j] = players[i].Cards[j].Points
			roomMsg.CardsSuits[3*i+j] = players[i].Cards[j].Suits
		}
	}

	return roomMsg
}
*/

func tableInfoDevlivery(delay time.Duration, ch chan RoomMsg) {
	var t [VOL_TABLE_MAX]*time.Timer
	// delayAuto := 6 * time.Second
	// var TableUsers [VOL_TABLE_MAX][TABLE_PLAYERS_MAX]rcvMessage
	// var nextPlayerMsg rcvMessage
	var nextPlayerMsg RoomMsg

	// TablesUsersMaps := make([]map[string]int, VOL_TABLE_MAX)
	for i := 0; i < VOL_TABLE_MAX; i++ {
		// TablesUsersMaps[i] = make(map[string]int, TABLE_PLAYERS_MAX)
		t[i] = time.NewTimer(delay)
	}

	for {
		select {
		case rcv := <-ch:
			// rcv = addCardsInfo(rcv)
			log.Println("S:", rcv)
			nextPlayerMsg = rcv

			nextPlayerMsg.SeatID = 1
			nextPlayerMsg.Status[nextPlayerMsg.SeatID] = "AUTO"
			t[rcv.TID].Reset(delay)

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
			log.Println("T0S:", nextPlayerMsg)
			// log.Println(nextPlayerMsg.Status[nextPlayerMsg.SeatID])
			if nextPlayerMsg.Status[nextPlayerMsg.SeatID] == "AUTO" {
				sendChan <- nextPlayerMsg
			}
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
<<<<<<< HEAD
=======

func receiveJsonHandler(connection *websocket.Conn) {
	// var rcv rcvMessage
	var roomMsg RoomMsg
	ch := make(chan RoomMsg)

	defer close(done)

	delay := 12 * time.Second
	go tableInfoDevlivery(delay, ch)

	for {
		err := connection.ReadJSON(&roomMsg)
		if err != nil {
			log.Println("Received not JSON data!")
			continue
		}
		log.Println("R:", roomMsg)
		ch <- roomMsg
	}
}
>>>>>>> d6972a01e97c63df6da4068c043b0596b05a2452
