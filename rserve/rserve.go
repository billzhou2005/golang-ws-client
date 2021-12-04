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
	Type       string    `json:"type"`
	RID        int       `json:"rID"`
	GameRound  int       `json:"gameRound"`
	BetRound   int       `json:"betRound"`
	DefendSeat int       `json:"defendSeat"`
	Focuses    [9]bool   `json:"focuses"`
	Players    [9]string `json:"players"`
	Balances   [9]int    `json:"balances"`
	CheckCards [9]bool   `json:"checkCards"`
	Discards   [9]bool   `json:"discards"`
	Reserve    string    `json:"reserve"`
}

type Player struct {
	Type      string `json:"type"`
	RID       int    `json:"rID"`
	PID       string `json:"pID"`
	MsgType   string `json:"msgType"`
	Name      string `json:"name"`
	GameRound int    `json:"gameRound"`
	BetRound  int    `json:"betRound"`
	SeatID    int    `json:"seatID"`
	SeatDID   int    `json:"seatDID"`
	Focus     bool   `json:"focus"`
	CheckCard bool   `json:"checkCard"`
	Discard   bool   `json:"discard"`
	BetVol    int    `json:"betVol"`
	Balance   int    `json:"balance"`
	Reserve   string `json:"reserve"`
}

type Cards struct {
	Type        string    `json:"type"`
	CardsName   string    `json:"cardsName"`
	RID         int       `json:"rID"`
	GameRound   int       `json:"gameRound"`
	CardsPoints [27]int   `json:"cardsPoints"`
	CardsSuits  [27]int   `json:"cardsSuits"`
	CardsTypes  [9]string `json:"cardsTypes"`
	Reserve     string    `json:"reserve"`
}

func init() {
	for i := 0; i < VOL_ROOM_MAX; i++ {
		Rooms[i].Type = "ROOM"
		Rooms[i].RID = 0
		Rooms[i].GameRound = 0
		Rooms[i].BetRound = 0
		Rooms[i].DefendSeat = 0
		Rooms[i].Reserve = "TBD"
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
			Rooms[i].Focuses[j] = false
			Rooms[i].Players[j] = "UNKNOWN"
			Rooms[i].Balances[j] = 0
			Rooms[i].CheckCards[j] = false
			Rooms[i].Discards[j] = true
		}
	}
}

// msgType: JOIN,ASSIGNED,NEWROUND,TIMEOUT,CLOSE, LEAVE

func RoomsUpdate(roomMsg RoomMsg) (time.Duration, bool, bool) {
	sendDelay := time.Millisecond
	msgDelivery := false
	cardsDelivery := false

	switch roomMsg.Type {
	case "JOIN":
		Rooms[roomMsg.RID] = addAutoPlayers(Rooms[roomMsg.RID])
		isOk := assignSeatID(roomMsg)
		sendDelay = time.Millisecond
		msgDelivery = true
		if !isOk {
			log.Println("login user assigned seat-failed")
			sendDelay = 3 * time.Second
		}
	case "ASSIGNED":
		Rooms[roomMsg.RID].Type = "NEWROUND"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "NEWROUND":
		Rooms[roomMsg.RID] = cardsShowSetFalse(Rooms[roomMsg.RID])
		Rooms[roomMsg.RID].Type = "CARDSDELIVERY"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "CARDSDELIVERY":
		cardsDelivery = true
		Rooms[roomMsg.RID] = cardsShowSetFalse(Rooms[roomMsg.RID])
		Rooms[roomMsg.RID].Type = "SETFOCUS"
		sendDelay = 1 * time.Second
		msgDelivery = false
	case "CHECKCARDS":
		sendDelay = time.Millisecond
		msgDelivery = true
		Rooms[roomMsg.RID] = playerCheckCards(Rooms[roomMsg.RID])
	case "SETFOCUS":
		sendDelay = time.Millisecond
		Rooms[roomMsg.RID].Type = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "WAITING":
		Rooms[roomMsg.RID].Type = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "CARDSCHECKED":
		sendDelay = time.Millisecond
		Rooms[roomMsg.RID].Type = "WAITING"
		msgDelivery = true
	case "LEAVE":
		sendDelay = 1 * time.Second
		msgDelivery = true
		Rooms[roomMsg.RID] = deleteLeavePlayers(Rooms[roomMsg.RID], roomMsg.Reserve)
		Rooms[roomMsg.RID].Type = "WAITING"
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
		if roomMsg.Players[i] != "UNKNOWN" {
			numofp++
		}
	}

	if numofp == 0 {
		for j := 0; j < 3; j++ {
			roomMsg.Players[j] = nickName[randomNums[j]]
			roomMsg.Focuses[j] = false
			roomMsg.CheckCards[j] = false
			roomMsg.Balances[j] = 6600000 + randomNums[j]*100000 // add random balance for auto user
		}
	}

	return roomMsg
}
func deleteLeavePlayers(roomMsg RoomMsg, name string) RoomMsg {

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Players[i] == name {
			roomMsg.Focuses[i] = false
			roomMsg.CheckCards[i] = false
			roomMsg.Players[i] = "UNKNOWN"
			roomMsg.Balances[i] = 0
		}
	}

	return roomMsg
}

func assignSeatID(roomMsg RoomMsg) bool {
	seatID := 100

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[roomMsg.RID].Players[i] == "UNKNOWN" {
			seatID = i
			break
		}
	}

	// check re-assigned or not
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[roomMsg.RID].Players[i] == roomMsg.Reserve {
			seatID = 100
			Rooms[roomMsg.RID].Type = "ASSIGNED"
			log.Println("Assgin SeatID failed, duplicated user:", roomMsg.Reserve)
			return false
		}
	}

	if seatID == 100 {
		log.Println("Assgin SeatID failed, the room is full:", ROOM_PLAYERS_MAX)
		return false
	}

	if seatID < ROOM_PLAYERS_MAX {
		Rooms[roomMsg.RID].Players[seatID] = roomMsg.Reserve
		Rooms[roomMsg.RID].Focuses[seatID] = false
		Rooms[roomMsg.RID].CheckCards[seatID] = false
		Rooms[roomMsg.RID].Balances[seatID] = 0

		Rooms[roomMsg.RID].Type = "ASSIGNED"
		Rooms[roomMsg.RID].Reserve = roomMsg.Reserve
	}

	return true
}

func playerCheckCards(roomMsg RoomMsg) RoomMsg {
	roomMsg.Type = "CARDSCHECKED"
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Reserve == roomMsg.Players[i] {
			roomMsg.CheckCards[i] = true
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
		roomMsg.CheckCards[i] = false
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
