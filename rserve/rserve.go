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
	Type        string `json:"type"`
	RID         int    `json:"rID"`
	Status      string `json:"status"`
	GameRound   int    `json:"gameRound"`
	BetRound    int    `json:"betRound"`
	FocusID     int    `json:"focusID"`
	BaseVol     int    `json:"baseVol"`
	TotalAmount int    `json:"totalAmount"`
	LostSeat    int    `json:"lostSeat"`
	WinnerSeat  int    `json:"winnerSeat"`
	DefendSeat  int    `json:"defendSeat"`
	Reserve     string `json:"reserve"`
}

type Player struct {
	Type      string       `json:"type"`
	RID       int          `json:"rID"`
	PID       string       `json:"pID"`
	MsgType   string       `json:"msgType"`
	Name      string       `json:"name"`
	SeatID    int          `json:"seatID"`
	SeatDID   int          `json:"seatDID"`
	BetRound  int          `json:"betRound"`
	Focus     bool         `json:"focus"`
	CheckCard bool         `json:"checkCard"`
	Discard   bool         `json:"discard"`
	BetVol    int          `json:"betVol"`
	Balance   int          `json:"balance"`
	Robot     bool         `json:"robot"`
	Cards     [3]util.Card `json:"Cards"`
	Reserve   string       `json:"reserve"`
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
		Rooms[i].RoomShare.BaseVol = 10000
		Rooms[i].RoomShare.DefendSeat = 0
		Rooms[i].RoomShare.Reserve = "TBD"
		for j := 0; j < ROOM_PLAYERS_MAX; j++ {
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

		sumNotDiscard := roomSumNotDiscard(room)
		if len(sumNotDiscard) == 2 {
			room.Players[sumNotDiscard[0]].Balance -= 100000
			room.RoomShare.TotalAmount += 100000
			room.Players[sumNotDiscard[1]].Balance -= 100000
			room.RoomShare.TotalAmount += 100000
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
		room.RoomShare.BetRound++
	case "SETTLE":
		sumNotDiscard := roomSumNotDiscard(room)
		if len(sumNotDiscard) == 1 {
			room.Players[sumNotDiscard[0]].Discard = true
			room.Players[sumNotDiscard[0]].MsgType = "WINNER"
			room.Players[sumNotDiscard[0]].Balance += room.RoomShare.TotalAmount
			room.RoomShare.TotalAmount = 0
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
	room.RoomsCards = util.GetPlayersCards(50000012, ROOM_PLAYERS_MAX)
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.Players[i].Name != "UNKNOWN" {
			if room.Players[i].Balance < room.RoomShare.BaseVol {
				room.Players[i].MsgType = "UNDERFUNDED"
				room.Players[i].Discard = true
			} else {
				room.Players[i].MsgType = "START"
				room.Players[i].Discard = false
				room.Players[i].Balance -= room.RoomShare.BaseVol
				room.RoomShare.TotalAmount += room.RoomShare.BaseVol
			}

			room.Players[i].Discard = false
			room.Players[i].Focus = false
			room.Players[i].CheckCard = false
			room.Players[i].Cards = room.RoomsCards[i].Cards
		}
	}

	room.RoomShare.Status = "START"
	room.RoomShare.BetRound = 0
	room.RoomShare.TotalAmount = 0
	room.RoomShare.DefendSeat = room.RoomShare.WinnerSeat
	room.RoomShare.FocusID = 0

	f := room.RoomShare.DefendSeat
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		f++
		if f >= ROOM_PLAYERS_MAX {
			f = 0
		}
		if room.Players[f].Name != "UNKNOWN" && !room.Players[i].Discard {
			room.Players[f].Focus = true
			room.RoomShare.FocusID = f
			break
		}
	}
	room.RoomShare.LostSeat = 100
	room.RoomShare.WinnerSeat = 100
	return room
}

func playerCardsCompare(room Room, seat1 int, seat2 int) int {
	if room.RoomsCards[seat1].Cardsscore < room.RoomsCards[seat2].Cardsscore {
		return seat1
	}
	return seat2
}

func PlayerInfoProcess(player Player) (bool, Player) {
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
		if room.Players[i].Name != "UNKNOWN" {
			numofp++
		}
	}

	if numofp == 0 {
		for j := 0; j < 3; j++ {
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
			room.Players[j].Balance = 500000 + randomNums[j]*100000 // add random balance for auto user
			room.Players[j].Robot = true
			room.Players[j].Reserve = "TBD"
		}
	}

	return room
}

/*
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
} */

func deleteLeavePlayers(room Room, player Player) Room {
	seatID := 100
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if room.Players[i].Name == player.Name {
			seatID = i
		}
	}

	if seatID > 8 {
		log.Println("Delete Player failed", seatID)
		return room
	}
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
		if Rooms[player.RID].Players[i].Name == "UNKNOWN" {
			seatID = i
			break
		}
	}

	// check re-assigned or not
	for i := 0; i < ROOM_PLAYERS_MAX; i++ {
		if Rooms[player.RID].Players[i].Name == player.Name {
			seatID = 100
			log.Println("Assgin SeatID failed, duplicated user:", player.Name)
			return false, seatID
		}
	}

	if seatID == 100 {
		log.Println("Assgin SeatID failed, the room is full:", ROOM_PLAYERS_MAX)
		return false, seatID
	}

	return true, seatID
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
