package rserve

import (
	"golang-ws-client/util"
	"log"
	"math/rand"
	"time"
)

const VOL_ROOM_MAX int = 3
const ROOM_PLAYERS_MAX int = 9

var Rooms [VOL_ROOM_MAX]RoomMsg

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

func init() {
	for i := 0; i < VOL_ROOM_MAX; i++ {
		Rooms[i].TID = 0
		Rooms[i].Name = "UNKNOWN"
		Rooms[i].MsgType = "NONE"
		Rooms[i].Reserve = "TBD"
		Rooms[i].SeatID = 0
		Rooms[i].Bvol = 0
		Rooms[i].Balance = 0
		Rooms[i].FID = 0
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
			Rooms[i].Focus[j] = false
			Rooms[i].CardsShow[j] = false
			Rooms[i].Names[j] = "UNKNOWN"
			Rooms[i].Balances[j] = 0
		}
	}
}

// msgType: JOIN,ASSIGNED,NEWROUND,TIMEOUT,CLOSE, LEAVE

func RoomsUpdate(roomMsg RoomMsg) (time.Duration, bool, bool) {
	sendDelay := time.Millisecond
	msgDelivery := false
	cardsDelivery := false

	switch roomMsg.MsgType {
	case "JOIN":
		Rooms[roomMsg.TID] = addAutoPlayers(Rooms[roomMsg.TID])
		isOk := assignSeatID(roomMsg)
		sendDelay = time.Millisecond
		msgDelivery = true
		if !isOk {
			log.Println("login user assigned seat-failed")
			sendDelay = 3 * time.Second
		}
	case "ASSIGNED":
		Rooms[roomMsg.TID].MsgType = "NEWROUND"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "NEWROUND":
		Rooms[roomMsg.TID] = cardsShowSetFalse(Rooms[roomMsg.TID])
		Rooms[roomMsg.TID].MsgType = "CARDSDELIVERY"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "CARDSDELIVERY":
		cardsDelivery = true
		Rooms[roomMsg.TID] = cardsShowSetFalse(Rooms[roomMsg.TID])
		Rooms[roomMsg.TID].MsgType = "SETFOCUS"
		sendDelay = 1 * time.Second
		msgDelivery = false
	case "CHECKCARDS":
		sendDelay = time.Millisecond
		msgDelivery = true
		Rooms[roomMsg.TID] = playerCheckCards(Rooms[roomMsg.TID])
	case "SETFOCUS":
		sendDelay = time.Millisecond
		Rooms[roomMsg.TID].MsgType = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "WAITING":
		Rooms[roomMsg.TID].MsgType = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "CARDSCHECKED":
		sendDelay = time.Millisecond
		Rooms[roomMsg.TID].MsgType = "WAITING"
		msgDelivery = true
	case "LEAVE":
		sendDelay = 1 * time.Second
		msgDelivery = true
		Rooms[roomMsg.TID] = deleteLeavePlayers(Rooms[roomMsg.TID], roomMsg.Name)
		Rooms[roomMsg.TID].MsgType = "WAITING"
	default:
		log.Println("Rooms info no need to update")
		sendDelay = 12 * time.Second
		msgDelivery = true
	}

	return sendDelay, msgDelivery, cardsDelivery
}

func AddCardsInfo(cards Cards) Cards {
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
		if Rooms[roomMsg.TID].Names[i] == "UNKNOWN" {
			seatID = i
			break
		}
	}

	// check re-assigned or not
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[roomMsg.TID].Names[i] == roomMsg.Name {
			seatID = 100
			Rooms[roomMsg.TID].MsgType = "ASSIGNED"
			log.Println("Assgin SeatID failed, duplicated user:", roomMsg.Name)
			return false
		}
	}

	if seatID == 100 {
		log.Println("Assgin SeatID failed, the room is full:", ROOM_PLAYERS_MAX)
		return false
	}

	if seatID < ROOM_PLAYERS_MAX {
		Rooms[roomMsg.TID].Names[seatID] = roomMsg.Name
		Rooms[roomMsg.TID].Focus[seatID] = false
		Rooms[roomMsg.TID].CardsShow[seatID] = false
		Rooms[roomMsg.TID].Balances[seatID] = roomMsg.Balance

		Rooms[roomMsg.TID].MsgType = "ASSIGNED"
		Rooms[roomMsg.TID].Name = roomMsg.Name
		Rooms[roomMsg.TID].SeatID = seatID
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

func generateRandomNumber(start int, end int, count int) []int {
	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn((end - start)) + start

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
