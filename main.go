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
	Types    [9]string `json:"types"`
	Names    [9]string `json:"names"`
	Balances [9]int    `json:"balances"`
}

type Cards struct {
	CardsPoints [27]int   `json:"cardsPoints"`
	CardsSuits  [27]int   `json:"cardsSuits"`
	Cardstypes  [9]string `json:"cardstypes"`
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
	sendChan = make(chan map[string]interface{})
	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	// rooms info init
	for i := 0; i < VOL_ROOM_MAX; i++ {
		rooms[i].TID = 0
		rooms[i].Name = "UNKNOWN"
		rooms[i].MsgType = "NONE"
		rooms[i].Type = "NONE"
		rooms[i].SeatID = 0
		rooms[i].Bvol = 0
		rooms[i].Balance = 0
		rooms[i].FID = 0
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
			rooms[i].Status[j] = "MANUAL"
			rooms[i].Types[j] = "NONE"
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
		log.Println("R:", rcvMsg)
		ch <- rcvMsg

		/* delete map after used
		for k := range rcvMsg {
			delete(rcvMsg, k)
		} */

	}
}

func addCardsInfo(cards Cards) Cards {
	players := util.GetPlayersCards(50000012, 9)
	// fmt.Println(players)

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		cards.Cardstypes[i] = players[i].Cardstype
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
		log.Println("msgType not found, return empty data")
		return roomMsg, false
	}

	return roomMsg, true
}

func addAutoPlayers(roomMsg RoomMsg) RoomMsg {
	var nickName = [...]string{"流逝的风", "每天赢5千", "不好就不要", "牛牛牛009", "风清猪的", "总是输没完了", "适度就是", "无畏了吗", "见好就收", "坚持到底", "三手要比", "不勉强", "搞不懂", "吴潇无暇", "大赌棍", "一直无感", "逍遥子", "风月浪", "独善其身", "赌神"}
	var numofp int

	randomNums := generateRandomNumber(0, 19, 3)

	numofp = 0
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Names[0] != "UNKNOWN" {
			numofp++
		}
	}

	if numofp == 0 {
		for j := 0; j < 3; j++ {
			roomMsg.Names[j] = nickName[randomNums[j]]
			roomMsg.Status[j] = "AUTO"
			roomMsg.Types[j] = "ASSIGNED"
			roomMsg.Balances[j] = 6600000 + randomNums[j]*100000 // add random balance for auto user
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
		rooms[roomMsg.TID].Status[seatID] = "MANUAL"
		rooms[roomMsg.TID].Types[seatID] = "ASSIGNED"
		rooms[roomMsg.TID].Balances[seatID] = roomMsg.Balance

		rooms[roomMsg.TID].MsgType = "ASSIGNED"
		rooms[roomMsg.TID].Name = roomMsg.Name
		rooms[roomMsg.TID].SeatID = seatID
	}

	return true
}

// msgType: JOIN,ASSIGNED,TIMEOUT,CLOSE

func roomsUpdate(roomMsg RoomMsg) {

	switch roomMsg.MsgType {
	case "JOIN":
		rooms[roomMsg.TID] = addAutoPlayers(rooms[roomMsg.TID])
		isOk := assignSeatID(roomMsg)
		if !isOk {
			log.Println("login user assigned seat-failed")
		}
	default:
		log.Println("rooms info no need to update")
	}
}

func tableInfoDevlivery(delay time.Duration, ch chan map[string]interface{}) {
	var t [VOL_ROOM_MAX]*time.Timer
	// delayAuto := 6 * time.Second
	// var nextPlayerMsg rcvMessage
	var roomNextMsg RoomMsg
	var cards Cards

	sendMap := make(map[string]interface{})

	testIndex := 0

	for i := 0; i < VOL_ROOM_MAX; i++ {
		t[i] = time.NewTimer(delay)
	}
	// roomNextMsg init
	roomNextMsg = rooms[0]

	for {
		select {
		case rcv := <-ch:
			// test code
			testIndex++
			cards = addCardsInfo(cards)

			rcvMsg, convertFlag := mapToStructRoomMsg(rcv)
			if convertFlag {
				roomsUpdate(rcvMsg)
			}

			roomNextMsg = rooms[rcvMsg.TID]

			t[0].Reset(delay)
			continue
		case <-t[0].C:

			// delete map before assigned
			for k := range sendMap {
				delete(sendMap, k)
			}

			log.Println("textIndex", testIndex)
			if testIndex%2 == 0 {
				msgMapSend(cardsStructToMap(cards))
				// sendMap = cardsStructToMap(cards)
			} else {
				msgMapSend(roomMsgStructToMap(roomNextMsg))
				// sendMap = roomMsgStructToMap(roomNextMsg)
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
>>>>>>> d6972a01e97c63df6da4068c043b0596b05a2452
