package rserve

import (
	"golang-ws-client/util"
	"log"
	"math/rand"
	"time"
)

type Err struct {
	code int
	des  string
}

const VOL_ROOM_MAX int = 3
const ROOM_PLAYERS_MAX int = 9

var Rooms [VOL_ROOM_MAX]RoomMsg

type RoomMsg struct {
	Type       string    `json:"type"`
	RID        int       `json:"rID"`
	Status     string    `json:"status"`
	GameRound  int       `json:"gameRound"`
	BetRound   int       `json:"betRound"`
	DefendSeat int       `json:"defendSeat"`
	Focuses    [9]bool   `json:"focuses"`
	Players    [9]string `json:"players"`
	Balances   [9]int    `json:"balances"`
	CheckCards [9]bool   `json:"checkCards"`
	Discards   [9]bool   `json:"discards"`
	Robots     [9]bool   `json:"robots"`
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
	Robot     bool   `json:"robot"`
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
		Rooms[i].Status = "WAITING"
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
			Rooms[i].Robots[j] = false
		}
	}
	Rooms[0] = addAutoPlayers(Rooms[0])
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
			roomMsg.Focuses[j] = false
			roomMsg.Players[j] = nickName[randomNums[j]]
			roomMsg.Balances[j] = 6600000 + randomNums[j]*100000 // add random balance for auto user
			roomMsg.CheckCards[j] = false
			roomMsg.Discards[j] = false
			roomMsg.Robots[j] = true
		}
	}

	return roomMsg
}

func RoomStartSet(roomMsg RoomMsg) RoomMsg {
	if roomMsg.Status != "WAITING" {
		return roomMsg
	}

	numofp := 0
	focusCount := 0
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if roomMsg.Players[i] != "UNKNOWN" {
			numofp++
			if roomMsg.Focuses[i] {
				focusCount++
			}
		}
	}
	if numofp < 1 {
		roomMsg.Status = "WAITING"
		return roomMsg
	}

	switch focusCount {
	case 0:
		for i := 0; i < ROOM_PLAYERS_MAX; i++ {
			if !roomMsg.Focuses[i] && roomMsg.Players[i] != "UNKNOWN" {
				roomMsg.Focuses[i] = true
				break
			}
		}
	case 1:
		log.Println("Focuses set correctly:", focusCount)
	default:
		log.Println("focusCount Invalid error:", focusCount)
	}
	roomMsg.Status = "START"
	return roomMsg
}

// msgType: JOIN,ASSIGNED,NEWROUND,TIMEOUT,CLOSE, LEAVE

func RoomsUpdate(roomMsg RoomMsg) (time.Duration, bool, bool) {
	sendDelay := time.Millisecond
	msgDelivery := false
	cardsDelivery := false

	switch roomMsg.Status {
	case "START":
		log.Println("Received Room data!-3", roomMsg.Status)
		Rooms[roomMsg.RID] = cardsShowSetFalse(Rooms[roomMsg.RID])
		Rooms[roomMsg.RID].Status = "CARDSDELIVERY"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "JOIN":
		isOk := assignSeatID(roomMsg)
		sendDelay = time.Millisecond
		msgDelivery = true
		if !isOk {
			log.Println("login user assigned seat-failed")
			sendDelay = 3 * time.Second
		}
	case "ASSIGNED":
		Rooms[roomMsg.RID].Status = "NEWROUND"
		sendDelay = 1 * time.Second
		msgDelivery = true
	case "CARDSDELIVERY":
		cardsDelivery = true
		Rooms[roomMsg.RID] = cardsShowSetFalse(Rooms[roomMsg.RID])
		Rooms[roomMsg.RID].Status = "SETFOCUS"
		sendDelay = 1 * time.Second
		msgDelivery = false
	case "CHECKCARDS":
		sendDelay = time.Millisecond
		msgDelivery = true
		Rooms[roomMsg.RID] = playerCheckCards(Rooms[roomMsg.RID])
	case "SETFOCUS":
		sendDelay = time.Millisecond
		Rooms[roomMsg.RID].Status = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "WAITING":
		Rooms[roomMsg.RID].Status = "WAITING"
		sendDelay = 12 * time.Second
		msgDelivery = true
	case "CARDSCHECKED":
		sendDelay = time.Millisecond
		Rooms[roomMsg.RID].Status = "WAITING"
		msgDelivery = true
	case "LEAVE":
		sendDelay = 1 * time.Second
		msgDelivery = true
		Rooms[roomMsg.RID] = deleteLeavePlayers(Rooms[roomMsg.RID], roomMsg.Reserve)
		Rooms[roomMsg.RID].Status = "WAITING"
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
			Rooms[roomMsg.RID].Status = "ASSIGNED"
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

		Rooms[roomMsg.RID].Status = "ASSIGNED"
		Rooms[roomMsg.RID].Reserve = roomMsg.Reserve
	}

	return true
}

func playerCheckCards(roomMsg RoomMsg) RoomMsg {
	roomMsg.Status = "CARDSCHECKED"
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
