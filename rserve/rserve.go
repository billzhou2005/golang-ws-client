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

type Room struct {
	Activated  bool                          `json:"activated"`
	RoomShare  RoomShare                     `json:"roomShare"`
	Players    [ROOM_PLAYERS_MAX]Player      `json:"players"`
	RoomsCards [ROOM_PLAYERS_MAX]util.Player `json:"roomCards"`
}

type RoomShare struct {
	Type       string                   `json:"type"`
	RID        int                      `json:"rID"`
	Status     string                   `json:"status"`
	GameRound  int                      `json:"gameRound"`
	BetRound   int                      `json:"betRound"`
	LostSeat   int                      `json:"lostSeat"`
	WinnerSeat int                      `json:"winnerSeat"`
	DefendSeat int                      `json:"defendSeat"`
	Focuses    [ROOM_PLAYERS_MAX]bool   `json:"focuses"`
	Players    [ROOM_PLAYERS_MAX]string `json:"players"`
	Balances   [ROOM_PLAYERS_MAX]int    `json:"balances"`
	Reserve    string                   `json:"reserve"`
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
	Type        string                    `json:"type"`
	CardsName   string                    `json:"cardsName"`
	RID         int                       `json:"rID"`
	GameRound   int                       `json:"gameRound"`
	CardsPoints [3 * ROOM_PLAYERS_MAX]int `json:"cardsPoints"`
	CardsSuits  [3 * ROOM_PLAYERS_MAX]int `json:"cardsSuits"`
	CardsTypes  [ROOM_PLAYERS_MAX]string  `json:"cardsTypes"`
	Reserve     string                    `json:"reserve"`
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
			Rooms[i].Players[j].RID = i
			Rooms[i].Players[j].Name = "UNKNOWN"
			Rooms[i].Players[j].Discard = true
		}
	}

	Rooms[0] = addAutoPlayers(Rooms[0])
}

func RoomStatusUpdate(room Room) Room {
	log.Println(room.RoomShare.Status)
	switch room.RoomShare.Status {
	case "WAITING":
		room = roomPlayersStartUpdate(room)
	case "START":
		room.RoomShare.Status = "BETTING"
	case "BETTING":
		room.RoomShare.BetRound++

		sumNotDiscard := roomSumNotDiscard(room)
		if len(sumNotDiscard) == 2 {
			lostSeat := playerCardsCompare(room, sumNotDiscard[0], sumNotDiscard[1])
			room.Players[lostSeat].Discard = true
			room.Players[lostSeat].MsgType = "LOST"
			room.RoomShare.LostSeat = lostSeat
		}
		if len(sumNotDiscard) == 1 {
			room.RoomShare.Status = "SETTLE"
		}
		/*
			b, _ := json.Marshal(room)
			log.Println("Room info:")
			log.Println(string(b)) */
		robotSeatIDs := roomSumRobostsNotDiscard(room)
		if len(robotSeatIDs) > 2 {
			lostSeat := playerCardsCompare(room, robotSeatIDs[0], robotSeatIDs[1])
			room.Players[lostSeat].Discard = true
			room.Players[lostSeat].MsgType = "LOST"
			room.RoomShare.LostSeat = lostSeat
			room.RoomShare.Status = "BETTING"
		}
	case "SETTLE":
		sumNotDiscard := roomSumNotDiscard(room)
		if len(sumNotDiscard) == 1 {
			room.Players[sumNotDiscard[0]].Discard = true
			room.Players[sumNotDiscard[0]].MsgType = "WINNER"
			room.RoomShare.WinnerSeat = sumNotDiscard[0]
		} else {
			room.RoomShare.Status = "WAITING"
			room.RoomShare.GameRound++
		}
	default:
		log.Println("Unknow Room Status", room.RoomShare.Status)
		room.RoomShare.Status = "WAITING"
	}
	return room
}

func roomSumRobostsNotDiscard(room Room) (seatIDs []int) {
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.Players[i].Robot && !room.Players[i].Discard {
			seatIDs = append(seatIDs, i)
		}
	}
	return seatIDs
}

func roomSumNotDiscard(room Room) (seatIDs []int) {
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if !room.Players[i].Discard && room.Players[i].Name != "UNKNOWN" {
			seatIDs = append(seatIDs, i)
		}
	}
	return seatIDs
}

func roomPlayersStartUpdate(room Room) Room {
	room.RoomShare.Status = "START"
	room.RoomShare.BetRound = 0
	room.RoomShare.DefendSeat = 0
	room.RoomShare.LostSeat = 100
	room.RoomShare.WinnerSeat = 100
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.RoomShare.Players[i] != "UNKNOWN" {
			room.Players[i].MsgType = "START"
			room.Players[i].Discard = false
			room.Players[i].Focus = false
			room.Players[i].CheckCard = false
		}
	}

	return room
}

func playerCardsCompare(room Room, seat1 int, seat2 int) int {
	if room.RoomsCards[seat1].Cardsscore < room.RoomsCards[seat2].Cardsscore {
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
	cards.RID = rID
	cards.CardsName = "jhCards"
	cards.GameRound = Rooms[rID].RoomShare.GameRound

	Rooms[rID].RoomsCards = util.GetPlayersCards(50000012, ROOM_PLAYERS_MAX)

	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		cards.CardsTypes[i] = Rooms[rID].RoomsCards[i].Cardstype
		for j := 0; j < 3; j++ {
			cards.CardsPoints[3*i+j] = Rooms[rID].RoomsCards[i].Cards[j].Points
			cards.CardsSuits[3*i+j] = Rooms[rID].RoomsCards[i].Cards[j].Suits
		}
	}

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
		if i == ROOM_PLAYERS_MAX {
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
