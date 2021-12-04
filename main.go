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

var cardsDelivery bool
var msgDelivery bool

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

	cardsDelivery = false
	msgDelivery = false

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

		roomMsg, cf1 := mapToStructRoomMsg(rcvMsg)
		cards, cf2 := mapToStructCards(rcvMsg)

		if cf1 {
			log.Println("R-roomMsg:", roomMsg)
			ch <- roomMsgStructToMap(roomMsg)
		}
		if cf2 {
			log.Println("R-cards:", cards)
			ch <- cardsStructToMap(cards)
		}

		if !cf1 && !cf2 {
			log.Println("Not roomMsg or cards, invalid message from readJSON")
		}

		// delete map after used
		for k := range rcvMsg {
			delete(rcvMsg, k)
		}

	}
}

func tableInfoDevlivery(delay time.Duration, ch chan map[string]interface{}) {
	var t [rserve.VOL_ROOM_MAX]*time.Timer
	// delayAuto := 6 * time.Second
	// var nextPlayerMsg rcvMessage
	var roomNextMsg rserve.RoomMsg
	var cards rserve.Cards
	var sendDelay time.Duration

	sendMap := make(map[string]interface{})

	for i := 0; i < rserve.VOL_ROOM_MAX; i++ {
		t[i] = time.NewTimer(delay)
	}
	// roomNextMsg init
	roomNextMsg = rserve.Rooms[0]

	for {
		select {
		case rcv := <-ch:

			rcvMsg, convertFlag := mapToStructRoomMsg(rcv)
			if convertFlag {
				log.Println("rcvMsg:", rcvMsg)
				sendDelay, msgDelivery, cardsDelivery = rserve.RoomsUpdate(rcvMsg)
			}

			roomNextMsg = rserve.Rooms[rcvMsg.TID]

			log.Println("msgDelivery:", msgDelivery, "cardsDelivery:", cardsDelivery, "sendDelay:", sendDelay)

			t[0].Reset(sendDelay)

			continue
		case <-t[0].C:
			fmt.Println("T0 response---")

			// delete map before assigned
			for k := range sendMap {
				delete(sendMap, k)
			}

			if (!cardsDelivery && !msgDelivery) && roomNextMsg.MsgType == "SETFOCUS" {
				msgMapSend(roomMsgStructToMap(roomNextMsg))
			}

			if cardsDelivery {
				cards = rserve.AddCardsInfo(cards)
				cards.TID = 0
				msgMapSend(cardsStructToMap(cards))
				cardsDelivery = false
			}
			if msgDelivery {
				msgMapSend(roomMsgStructToMap(roomNextMsg))
				msgDelivery = false
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

func msgMapSend(msgMap map[string]interface{}) {
	sendMap := make(map[string]interface{})
	for k := range sendMap {
		delete(sendMap, k)
	}
	sendMap = msgMap
	sendChan <- sendMap
}

func mapToStructRoomMsg(m map[string]interface{}) (rserve.RoomMsg, bool) {
	var roomMsg rserve.RoomMsg

	_, isOk := m["msgType"]
	if isOk {
		arr, err := json.Marshal(m)
		if err != nil {
			fmt.Println(err)
			return roomMsg, false
		}
		err = json.Unmarshal(arr, &roomMsg)
		if err != nil {
			fmt.Println(err)
			return roomMsg, false
		}
	} else {
		log.Println("RoomMsg struct not found, return empty data")
		return roomMsg, false
	}

	return roomMsg, true
}

func mapToStructCards(m map[string]interface{}) (rserve.Cards, bool) {
	var cards rserve.Cards

	_, isOk := m["cardsTypes"]
	if isOk {
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
	} else {
		log.Println("Cards struct not found, return empty data")
		return cards, false
	}

	return cards, true
}
func roomMsgStructToMap(roomMsg rserve.RoomMsg) map[string]interface{} {
	tempMap := make(map[string]interface{})

	tmprec, _ := json.Marshal(&roomMsg)
	json.Unmarshal(tmprec, &tempMap)

	return tempMap
}

func cardsStructToMap(cards rserve.Cards) map[string]interface{} {
	tempMap := make(map[string]interface{})

	tmprec, _ := json.Marshal(&cards)
	json.Unmarshal(tmprec, &tempMap)

	return tempMap
}
