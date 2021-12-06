// websocket client.go
package main

import (
	"encoding/json"
	"fmt"
	"golang-ws-client/rserve"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var done chan interface{}
var interrupt chan os.Signal
var sendChan chan map[string]interface{}

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
		fmt.Println(msg)
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

	go receiveJsonHandler(conn)

	for {
		select {
		case sendMsg := <-sendChan:

			err := conn.WriteJSON(sendMsg)
			if err != nil {
				log.Println("Error during writing the json data to websocket:", err)
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

func receiveJsonHandler(connection *websocket.Conn) {
	var rcvMsg map[string]interface{}

	chPlayer := make(chan rserve.Player)

	defer close(done)

	go roomServe(chPlayer)

	for {
		err := connection.ReadJSON(&rcvMsg)
		if err != nil {
			log.Println("Received not JSON data!")
			continue
		}

		switch rcvMsg["type"] {
		case "PLAYER":
			player, isOk := mapToStructPlayer(rcvMsg)
			if isOk {
				chPlayer <- player
			} else {
				log.Println("Received Player Struct error")
			}
		case "ROOM":
			room, isOk := mapToStructRoomMsg(rcvMsg)
			log.Println("RoomShare is:", isOk, room)
		case "CARDS":
			cards, isOk := mapToStructCards(rcvMsg)
			log.Println("Cards is:", isOk, cards)
		default:
			log.Println("Not room/player/cards json", rcvMsg)
		}

		// Empty map
		for k := range rcvMsg {
			delete(rcvMsg, k)
		}

	}
}

func roomServe(chPlayer chan rserve.Player) {
	var t [rserve.VOL_ROOM_MAX]*time.Timer
	var cards rserve.Cards

	delay := 6 * time.Second

	for i := 0; i < rserve.VOL_ROOM_MAX; i++ {
		t[i] = time.NewTimer(delay)
	}

	for {
		select {
		case player := <-chPlayer:

			log.Println("roomServe", player)
			isOk, player := rserve.PlayerInfoProcess(player)
			if isOk {
				msgMapSend(playerStructToMap(player))
			}

			continue
		case <-t[0].C:
			rID := 0
			rserve.Rooms[rID] = rserve.RoomStatusUpdate(rserve.Rooms[rID])
			msgMapSend(roomShareStructToMap(rserve.Rooms[rID].RoomShare))

			if rserve.Rooms[rID].RoomShare.Status == "START" {
				cards = rserve.AddCardsInfo(cards, rserve.Rooms[rID].RoomShare.RID)
				msgMapSend(cardsStructToMap(cards))
			}
			if rserve.Rooms[rID].RoomShare.Status == "BETTING" {
				if rserve.Rooms[rID].RoomShare.LostSeat < rserve.ROOM_PLAYERS_MAX {
					<-time.After(time.Second * 1)
					msgMapSend(playerStructToMap(rserve.Rooms[rID].Players[rserve.Rooms[rID].RoomShare.LostSeat]))
					rserve.Rooms[rID].RoomShare.LostSeat = 100 // reset
				}
			}
			if rserve.Rooms[rID].RoomShare.Status == "SETTLE" {
				if rserve.Rooms[rID].RoomShare.WinnerSeat < rserve.ROOM_PLAYERS_MAX {
					<-time.After(time.Second * 1)
					msgMapSend(playerStructToMap(rserve.Rooms[rID].Players[rserve.Rooms[rID].RoomShare.WinnerSeat]))
					rserve.Rooms[rID].RoomShare.WinnerSeat = 100 // reset
				}
			}

			log.Println("T0 message, interval:", delay)
			t[rID].Reset(delay)
		case <-t[1].C:
			log.Println("T1 message, interval:", delay)
			// t[1].Reset(delay)
		case <-t[2].C:
			log.Println("T2 message, interval:", delay)
			// t[2].Reset(delay)
		}
	}
}

func msgMapSend(msgMap map[string]interface{}) {
	sendMap := make(map[string]interface{})
	for k := range sendMap {
		delete(sendMap, k)
	}
	sendMap = msgMap
	sendChan <- sendMap
}

func mapToStructRoomMsg(m map[string]interface{}) (rserve.RoomShare, bool) {
	var room rserve.RoomShare

	arr, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return room, false
	}
	err = json.Unmarshal(arr, &room)
	if err != nil {
		fmt.Println(err)
		return room, false
	}

	return room, true
}

func mapToStructPlayer(m map[string]interface{}) (rserve.Player, bool) {
	var player rserve.Player

	arr, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return player, false
	}
	err = json.Unmarshal(arr, &player)
	if err != nil {
		fmt.Println(err)
		return player, false
	}

	return player, true
}

func mapToStructCards(m map[string]interface{}) (rserve.Cards, bool) {
	var cards rserve.Cards

	arr, err := json.Marshal(m)
	if err != nil {
		fmt.Println(err)
		return cards, false
	}
	err = json.Unmarshal(arr, &cards)
	if err != nil {
		fmt.Println(err)
		return cards, false
	}

	return cards, true
}

func roomShareStructToMap(roomShare rserve.RoomShare) map[string]interface{} {
	tempMap := make(map[string]interface{})

	temp, _ := json.Marshal(&roomShare)
	json.Unmarshal(temp, &tempMap)

	return tempMap
}

func playerStructToMap(player rserve.Player) map[string]interface{} {
	tempMap := make(map[string]interface{})

	temp, _ := json.Marshal(&player)
	json.Unmarshal(temp, &tempMap)

	return tempMap
}

func cardsStructToMap(cards rserve.Cards) map[string]interface{} {
	tempMap := make(map[string]interface{})

	temp, _ := json.Marshal(&cards)
	json.Unmarshal(temp, &tempMap)

	return tempMap
}
