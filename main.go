// websocket client.go
package main

import (
	"encoding/json"
	"fmt"
	"golang-ws-client/util"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const VOL_ROOM_MAX int = 3
const ROOM_PLAYERS_MAX int = 9

var rooms [VOL_ROOM_MAX]RoomMsg
var cardsDelivery bool
var msgDelivery bool

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
	TID       int       `json:"tID"`
	Name      string    `json:"name"`
	MsgType   string    `json:"msgType"`
	Reserve   string    `json:"reserve"`
	SeatID    int       `json:"seatID"`
	Bvol      int       `json:"bvol"`
	Balance   int       `json:"balance"`
	FID       int       `json:"fID"`
	Focus     [9]bool   `json:"focus"`
	CardsShow [9]bool   `json:"cardsShow"`
	Names     [9]string `json:"names"`
	Balances  [9]int    `json:"balances"`
}

type Cards struct {
	TID         int       `json:"tID"`
	CardsName   string    `json:"cardsName"`
	CardsPoints [27]int   `json:"cardsPoints"`
	CardsSuits  [27]int   `json:"cardsSuits"`
	CardsTypes  [9]string `json:"cardsTypes"`
}

type Player struct {
	NickName string `json:"nickName"`
	Sts      string `json:"sts"`
	SID      int    `json:"sID"`
	Vol      int    `json:"vol"`
	Tol      int    `json:"Tol"`
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

	// rooms info init
	for i := 0; i < VOL_ROOM_MAX; i++ {
		rooms[i].TID = 0
		rooms[i].Name = "UNKNOWN"
		rooms[i].MsgType = "NONE"
		rooms[i].Reserve = "TBD"
		rooms[i].SeatID = 0
		rooms[i].Bvol = 0
		rooms[i].Balance = 0
		rooms[i].FID = 0
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
			rooms[i].Focus[j] = false
			rooms[i].CardsShow[j] = false
			rooms[i].Names[j] = "UNKNOWN"
			rooms[i].Balances[j] = 0
		}
	}

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

func addCardsInfo(cards Cards) Cards {
	cards.CardsName = "jhCards"

	players := util.GetPlayersCards(50000012, 9)
	// fmt.Println(players)

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		cards.CardsTypes[i] = players[i].Cardstype
		for j := 0; j < 3; j++ {
			cards.CardsPoints[3*i+j] = players[i].Cards[j].Points
			cards.CardsSuits[3*i+j] = players[i].Cards[j].Suits
		}
	}

	return cards
}

func mapToStructRoomMsg(m map[string]interface{}) (RoomMsg, bool) {
	var roomMsg RoomMsg

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

func mapToStructCards(m map[string]interface{}) (Cards, bool) {
	var cards Cards

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

func addAutoPlayers(roomMsg RoomMsg) RoomMsg {
	var nickName = [...]string{"流逝的风", "每天赢5千", "不好就不要", "牛牛牛009", "风清猪的", "总是输没完了", "适度就是", "无畏了吗", "见好就收", "坚持到底", "三手要比", "不勉强", "搞不懂", "吴潇无暇", "大赌棍", "一直无感", "逍遥子", "风月浪", "独善其身", "赌神"}
	var numofp int

	randomNums := generateRandomNumber(0, 19, 3)

	numofp = 0
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Names[i] != "UNKNOWN" {
			numofp++
		}
	}

	if numofp == 0 {
		for j := 0; j < 3; j++ {
			roomMsg.Names[j] = nickName[randomNums[j]]
			roomMsg.Focus[j] = false
			roomMsg.CardsShow[j] = false
			roomMsg.Balances[j] = 6600000 + randomNums[j]*100000 // add random balance for auto user
		}
	}

	return roomMsg
}
func deleteLeavePlayers(roomMsg RoomMsg, name string) RoomMsg {

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Names[i] == name {
			roomMsg.Focus[i] = false
			roomMsg.CardsShow[i] = false
			roomMsg.Names[i] = "UNKNOWN"
			roomMsg.Balances[i] = 0
		}
	}

	return roomMsg
}

func assignSeatID(roomMsg RoomMsg) bool {
	seatID := 100

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if rooms[roomMsg.TID].Names[i] == "UNKNOWN" {
			seatID = i
			break
		}
	}

	// check re-assigned or not
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if rooms[roomMsg.TID].Names[i] == roomMsg.Name {
			seatID = 100
			rooms[roomMsg.TID].MsgType = "ASSIGNED"
			log.Println("Assgin SeatID failed, duplicated user:", roomMsg.Name)
			return false
		}
	}

	if seatID == 100 {
		log.Println("Assgin SeatID failed, the room is full:", ROOM_PLAYERS_MAX)
		return false
	}

	if seatID < ROOM_PLAYERS_MAX {
		rooms[roomMsg.TID].Names[seatID] = roomMsg.Name
		rooms[roomMsg.TID].Focus[seatID] = false
		rooms[roomMsg.TID].CardsShow[seatID] = false
		rooms[roomMsg.TID].Balances[seatID] = roomMsg.Balance

		rooms[roomMsg.TID].MsgType = "ASSIGNED"
		rooms[roomMsg.TID].Name = roomMsg.Name
		rooms[roomMsg.TID].SeatID = seatID
	}

	return true
}

func playerCheckCards(roomMsg RoomMsg) RoomMsg {
	roomMsg.MsgType = "CARDSCHECKED"
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Name == roomMsg.Names[i] {
			roomMsg.CardsShow[i] = true
			break
		}
		if i == 9 {
			log.Println("Player not found in func: playerCheckCards")
		}
	}

	return roomMsg
}

func cardsShowSetFalse(roomMsg RoomMsg) RoomMsg {
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		roomMsg.CardsShow[i] = false
	}
	return roomMsg
}

// msgType: JOIN,ASSIGNED,NEWROUND,TIMEOUT,CLOSE, LEAVE

func roomsUpdate(roomMsg RoomMsg) time.Duration {
	sendDelay := time.Millisecond

	switch roomMsg.MsgType {
	case "JOIN":
		rooms[roomMsg.TID] = addAutoPlayers(rooms[roomMsg.TID])
		isOk := assignSeatID(roomMsg)
		sendDelay = time.Millisecond
		msgDelivery = true
		if !isOk {
			log.Println("login user assigned seat-failed")
			sendDelay = 3 * time.Second
		}
	case "ASSIGNED":
		rooms[roomMsg.TID].MsgType = "NEWROUND"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "NEWROUND":
		rooms[roomMsg.TID] = cardsShowSetFalse(rooms[roomMsg.TID])
		rooms[roomMsg.TID].MsgType = "CARDSDELIVERY"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "CARDSDELIVERY":
		cardsDelivery = true
		rooms[roomMsg.TID] = cardsShowSetFalse(rooms[roomMsg.TID])
		rooms[roomMsg.TID].MsgType = "SETFOCUS"
		sendDelay = 1 * time.Second
		msgDelivery = false
	case "CHECKCARDS":
		sendDelay = time.Millisecond
		msgDelivery = true
		rooms[roomMsg.TID] = playerCheckCards(rooms[roomMsg.TID])
	case "SETFOCUS":
		sendDelay = time.Millisecond
		rooms[roomMsg.TID].MsgType = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "WAITING":
		rooms[roomMsg.TID].MsgType = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "CARDSCHECKED":
		sendDelay = time.Millisecond
		rooms[roomMsg.TID].MsgType = "WAITING"
		msgDelivery = true
	case "LEAVE":
		sendDelay = 1 * time.Second
		msgDelivery = true
		rooms[roomMsg.TID] = deleteLeavePlayers(rooms[roomMsg.TID], roomMsg.Name)
		rooms[roomMsg.TID].MsgType = "WAITING"
	default:
		log.Println("rooms info no need to update")
		sendDelay = 12 * time.Second
		msgDelivery = true
	}

	return sendDelay
}

func tableInfoDevlivery(delay time.Duration, ch chan map[string]interface{}) {
	var t [VOL_ROOM_MAX]*time.Timer
	// delayAuto := 6 * time.Second
	// var nextPlayerMsg rcvMessage
	var roomNextMsg RoomMsg
	var cards Cards
	var sendDelay time.Duration

	sendMap := make(map[string]interface{})

	for i := 0; i < VOL_ROOM_MAX; i++ {
		t[i] = time.NewTimer(delay)
	}
	// roomNextMsg init
	roomNextMsg = rooms[0]

	for {
		select {
		case rcv := <-ch:

			rcvMsg, convertFlag := mapToStructRoomMsg(rcv)
			if convertFlag {
				log.Println("rcvMsg:", rcvMsg)
				sendDelay = roomsUpdate(rcvMsg)
			}

			roomNextMsg = rooms[rcvMsg.TID]

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
				cards = addCardsInfo(cards)
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

func roomMsgStructToMap(roomMsg RoomMsg) map[string]interface{} {
	tempMap := make(map[string]interface{})

	tmprec, _ := json.Marshal(&roomMsg)
	json.Unmarshal(tmprec, &tempMap)

	return tempMap
}

func cardsStructToMap(cards Cards) map[string]interface{} {
	tempMap := make(map[string]interface{})

	tmprec, _ := json.Marshal(&cards)
	json.Unmarshal(tmprec, &tempMap)

	return tempMap
}

//生成count个[start,end)结束的不重复的随机数
func generateRandomNumber(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start) < count {
		return nil
	}

	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}
