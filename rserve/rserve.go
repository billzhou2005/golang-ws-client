package rserve

import (
	"golang-ws-client/util"
	"log"
	"math/rand"
	"time"
)

const VOL_ROOM_MAX int = 3
const ROOM_PLAYERS_MAX int = 9

var Rooms [VOL_ROOM_MAX]Room
var RoomsCards [VOL_ROOM_MAX][ROOM_PLAYERS_MAX]util.Player

type Room struct {
	Activated bool      `json:"activated"`
	RoomShare RoomShare `json:"roomShare"`
	Players   [9]Player `json:"players"`
}

type RoomShare struct {
	Type       string    `json:"type"`
	RID        int       `json:"rID"`
	Status     string    `json:"status"`
	GameRound  int       `json:"gameRound"`
	BetRound   int       `json:"betRound"`
	DefendSeat int       `json:"defendSeat"`
	Focuses    [9]bool   `json:"focuses"`
	Players    [9]string `json:"players"`
	Balances   [9]int    `json:"balances"`
	Reserve    string    `json:"reserve"`
}

type Player struct {
	Type      string `json:"type"`
	RID       int    `json:"rID"`
	PID       string `json:"pID"`
	MsgType   string `json:"msgType"`
	Name      string `json:"name"`
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
		Rooms[i].Activated = true
		Rooms[i].RoomShare.Type = "ROOM"
		Rooms[i].RoomShare.RID = i
		Rooms[i].RoomShare.Status = "WAITING"
		Rooms[i].RoomShare.GameRound = 0
		Rooms[i].RoomShare.BetRound = 0
		Rooms[i].RoomShare.DefendSeat = 0
		Rooms[i].RoomShare.Reserve = "TBD"
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
			Rooms[i].RoomShare.Focuses[j] = false
			Rooms[i].RoomShare.Players[j] = "UNKNOWN"
			Rooms[i].RoomShare.Balances[j] = 0
		}
	}

	Rooms[0] = addAutoPlayers(Rooms[0])
}

func RoomStatusUpdate(room Room) Room {
	switch room.RoomShare.Status {
	case "WAITING":
		room.RoomShare.Status = "START"
	case "START":
		room.RoomShare.Status = "BETTING"
		room.RoomShare.GameRound++
	case "BETTING":
		room.RoomShare.Status = "SETTLE"

	case "SETTLE":
		playerCardsCompare(room, 0, 1)
		room.RoomShare.Status = "WAITING"
	default:
		log.Println("Unknow Room Status", room.RoomShare.Status)
		room.RoomShare.Status = "WAITING"
	}
	return room
}

func playerCardsCompare(room Room, seat1 int, seat2 int) int {
	log.Println("SeatID:", seat1, RoomsCards[room.RoomShare.RID][seat1].Cardsscore, RoomsCards[room.RoomShare.RID][seat1].Cards)
	log.Println("SeatID:", seat2, RoomsCards[room.RoomShare.RID][seat2].Cardsscore, RoomsCards[room.RoomShare.RID][seat2].Cards)
	if RoomsCards[room.RoomShare.RID][seat1].Cardsscore > RoomsCards[room.RoomShare.RID][seat2].Cardsscore {
		return seat1
	}
	return seat2
}

func PlayerInfoProcess(player Player) (bool, Player) {
	// RoomStartSet(Rooms[player.RID])
	switch player.MsgType {
	case "JOIN":
		isOk, seatID := assignSeatID(player)
		if isOk {
			player.SeatID = seatID
			player.MsgType = "ASSIGNED"
			Rooms[player.RID].Players[seatID] = player
			log.Println("login user assigned seat-Sucess")
			return true, player
		}
		log.Println(Rooms[player.RID])
	case "LEAVE":
		Rooms[player.RID] = deleteLeavePlayers(Rooms[player.RID], player)
	default:
		log.Println("player.MsgType", player.MsgType)

	}

	return false, player
}

func addAutoPlayers(room Room) Room {
	var nickName = [...]string{"流逝的风", "每天赢5千", "不好就不要", "牛牛牛009", "风清猪的", "总是输没完了", "适度就是", "无畏了吗", "见好就收", "坚持到底", "三手要比", "不勉强", "搞不懂", "吴潇无暇", "大赌棍", "一直无感", "逍遥子", "风月浪", "独善其身", "赌神"}
	var numofp int

	randomNums := generateRandomNumber(0, 19, 3)

	numofp = 0
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.RoomShare.Players[i] != "UNKNOWN" {
			numofp++
		}
	}

	if numofp == 0 {
		for j := 0; j < 3; j++ {
			room.RoomShare.Focuses[j] = false
			room.RoomShare.Players[j] = nickName[randomNums[j]]
			room.RoomShare.Balances[j] = 6600000 + randomNums[j]*100000 // add random balance for auto user
			room.Players[j].Type = "PLAYER"
			room.Players[j].RID = room.RoomShare.RID
			room.Players[j].PID = "xxxaaa88"
			room.Players[j].MsgType = "ASSIGNED"
			room.Players[j].Name = nickName[randomNums[j]]
			room.Players[j].SeatID = j
			room.Players[j].Focus = false
			room.Players[j].CheckCard = false
			room.Players[j].Discard = true
			room.Players[j].BetVol = 0
			room.Players[j].Balance = room.RoomShare.Balances[j]
			room.Players[j].Robot = true
			room.Players[j].Reserve = "TBD"
		}
	}

	return room
}

func RoomStartSet(room RoomShare) RoomShare {
	if room.Status != "WAITING" {
		return room
	}

	numofp := 0
	focusCount := 0
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.Players[i] != "UNKNOWN" {
			numofp++
			if room.Focuses[i] {
				focusCount++
			}
		}
	}
	if numofp < 1 {
		room.Status = "WAITING"
		return room
	}

	switch focusCount {
	case 0:
		for i := 0; i < ROOM_PLAYERS_MAX; i++ {
			if !room.Focuses[i] && room.Players[i] != "UNKNOWN" {
				room.Focuses[i] = true
				break
			}
		}
	case 1:
		log.Println("Focuses set correctly:", focusCount)
	default:
		log.Println("focusCount Invalid error:", focusCount)
	}
	room.Status = "START"
	return room
}

func AddCardsInfo(cards Cards, rID int) Cards {
	cards.Type = "CARDS"
	cards.CardsName = "jhCards"

	RoomsCards[rID] = util.GetPlayersCards(50000012, 9)

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		cards.CardsTypes[i] = RoomsCards[rID][i].Cardstype
		for j := 0; j < 3; j++ {
			cards.CardsPoints[3*i+j] = RoomsCards[rID][i].Cards[j].Points
			cards.CardsSuits[3*i+j] = RoomsCards[rID][i].Cards[j].Suits
		}
	}

	log.Println("Room cards update, rID:", rID, RoomsCards[rID])
	return cards
}

func deleteLeavePlayers(room Room, player Player) Room {
	seatID := 100
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.RoomShare.Players[i] == player.Name {
			seatID = i
		}
	}

	if seatID > 8 {
		log.Println("Delete Player failed", seatID)
		return room
	}
	room.RoomShare.Focuses[seatID] = false
	room.RoomShare.Players[seatID] = "UNKNOWN"
	room.RoomShare.Balances[seatID] = 0
	room.Players[seatID].Name = "UNKNOWN"
	room.Players[seatID].PID = "UNKNOWN"
	room.Players[seatID].MsgType = "LEFT"
	room.Players[seatID].SeatID = 100
	room.Players[seatID].SeatDID = 100
	room.Players[seatID].Focus = false
	room.Players[seatID].Discard = true
	room.Players[seatID].CheckCard = false
	room.Players[seatID].BetVol = 0
	room.Players[seatID].Balance = 0
	room.Players[seatID].Robot = false
	return room
}

func assignSeatID(player Player) (bool, int) {
	seatID := 100

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[player.RID].RoomShare.Players[i] == "UNKNOWN" {
			seatID = i
			break
		}
	}

	// check re-assigned or not
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[player.RID].RoomShare.Players[i] == player.Name {
			seatID = 100
			log.Println("Assgin SeatID failed, duplicated user:", player.Name)
			return false, seatID
		}
	}

	if seatID == 100 {
		log.Println("Assgin SeatID failed, the room is full:", ROOM_PLAYERS_MAX)
		return false, seatID
	}

	if seatID < ROOM_PLAYERS_MAX {
		Rooms[player.RID].RoomShare.Focuses[seatID] = false
		Rooms[player.RID].RoomShare.Players[seatID] = player.Name
		Rooms[player.RID].RoomShare.Balances[seatID] = player.Balance
	}

	return true, seatID
}

func playerCheckCards(room RoomShare) RoomShare {
	room.Status = "CARDSCHECKED"
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.Reserve == room.Players[i] {
			room.Focuses[i] = true
			break
		}
		if i == 9 {
			log.Println("Player not found in func: playerCheckCards")
		}
	}

	return room
}

func cardsShowSetFalse(room RoomShare) RoomShare {
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		room.Focuses[i] = false
	}
	return room
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
