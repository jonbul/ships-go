package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"ships/controllers/models"
)

type bulletData = models.BulletData
type wsEvent = models.WsEvent
type playerData = models.PlayerData
type playerHitData = models.PlayerHitData

type safeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (sc *safeConn) writeJSON(v any) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteJSON(v)
}

var mu sync.Mutex
var userConnections = make(map[string]*safeConn)
var playersToSend = make(map[string]*playerData)
var players = make(map[string]*playerData)
var newBullets = []*bulletData{}
var bulletsToRemove = []string{}
var killsList = []*playerHitData{}
var hasPlayersTosend = false

// backgroundCards
// {x: {y : [xInCard, yInCard,size(1 to 5)]}}
// {1: {1 : [1,2,3],2 : [1,2,3]}, 2: {1 : [1,2,3],2 : [1,2,3]...}...}

// var BackgroundCards = make(map[int]map[int][]any)
var BackgroundCards = make(map[int]map[int]Card)

type Card struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Stars []Star `json:"stars"`
}
type Star struct {
	X int `json:"x"`
	Y int `json:"y"`
	R int `json:"r"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func registerWebSocket(router *gin.Engine) {
	router.GET("/ws", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	go broadCastInterval()
	playersToSend = make(map[string]*playerData)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}

	socketId := uuid.New().String()
	defer func(conn *websocket.Conn, socketId string) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
		mu.Lock() // to ensure thread safety when modifying the maps
		delete(userConnections, socketId)
		delete(players, socketId)
		mu.Unlock()
	}(conn, socketId)
	sc := &safeConn{conn: conn}
	mu.Lock()
	userConnections[socketId] = sc
	mu.Unlock()
	// Listen for incoming messages
	for {
		// Read message from the client
		_, messagePlain, err := conn.ReadMessage()
		if err != nil {
			break
		}

		manageInputMessage(sc, messagePlain, socketId)
	}
}

func manageInputMessage(conn *safeConn, msgPlain []byte, socketId string) {
	var msg wsEvent
	_ = json.Unmarshal(msgPlain, &msg)

	switch msg.EventName {
	case "connectionSuccess":
		log.Println("New connection with socketId: " + socketId)
		msg.SocketId = socketId
		_ = conn.writeJSON(msg)
		return
	case "playerData":

		var plData *playerData
		_ = json.Unmarshal(msgPlain, &plData)
		if "" != socketId && "" != msg.SocketId {
			msg.SocketId = socketId
			mu.Lock()
			playersToSend[socketId] = plData
			players[socketId] = plData
			mu.Unlock()
		}
		return
	case "getBackgroundCards":
		wsGetBackgroundCards(conn)
		return
	case "newBullet":

		var bullet *bulletData
		_ = json.Unmarshal(msgPlain, &bullet)
		mu.Lock()
		newBullets = append(newBullets, bullet)
		mu.Unlock()
		return
	case "removeBullet":

		var playerHit *playerHitData
		_ = json.Unmarshal(msgPlain, &playerHit)
		mu.Lock()
		bulletsToRemove = append(bulletsToRemove, playerHit.BulletId)
		mu.Unlock()
		return
	case "playerHit":
		var playerHit *playerHitData
		_ = json.Unmarshal(msgPlain, &playerHit)
		mu.Lock()
		targetConn := userConnections[playerHit.PlayerId]
		mu.Unlock()
		if targetConn != nil {
			_ = targetConn.writeJSON(playerHit)
		}
		return
	case "playerDied":
		var plHitData *playerHitData
		_ = json.Unmarshal(msgPlain, &plHitData)
		mu.Lock()
		killsList = append(killsList, plHitData)
		playerFrom, ok := players[plHitData.From]
		if ok {
			hasPlayersTosend = true
			playerFrom.Credits += 100
		}
		mu.Unlock()
	default:
		log.Println("--------------------------")
		log.Println("Unknown event: " + msg.EventName)
		log.Println("--------------------------")
	}
}

func buildBackgroundCards() {
	if len(BackgroundCards) > 0 {
		return
	}
	var w = resolutions[currentResolution].Width
	var h = resolutions[currentResolution].Height

	for x := 0; x < 5; x++ {
		BackgroundCards[x] = make(map[int]Card)
		for y := 0; y < 5; y++ {

			var points [500]Star
			for i := 0; i < 500; i++ {
				star := Star{
					X: rand.IntN(w),
					Y: rand.IntN(h),
					R: rand.IntN(4) + 1,
				}
				points[i] = star
			}
			BackgroundCards[x][y] = Card{
				X:     x,
				Y:     y,
				Stars: points[:],
			}
		}
	}
}

func wsGetBackgroundCards(conn *safeConn) {
	buildBackgroundCards()
	_ = conn.writeJSON(gin.H{"eventName": "getBackgroundCards", "cards": BackgroundCards})
}

var lastBroadcastTime int64 = 0

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	for range ticker.C {
		ActivePlayers.Set(float64(len(players)))
		mu.Lock()

		var currentTime = time.Now().UnixMilli()

		// send at least every 2 seconds or if there are any new bullets, players, or kills to send
		if len(players) == 0 || (len(playersToSend)+len(newBullets)+len(killsList) == 0 && currentTime-lastBroadcastTime < 2000) {
			mu.Unlock()
			continue
		}
		lastBroadcastTime = currentTime

		var playerIds = make([]string, 0, len(players))
		for id := range players {
			playerIds = append(playerIds, id)
		}

		var payload = map[string]any{
			"eventName":       "gameBroadcast",
			"bulletsToRemove": bulletsToRemove,
			"newBullets":      newBullets,
			"players":         playersToSend,
			"kills":           killsList,
			"activePlayerIds": playerIds,
		}

		conns := make([]*safeConn, 0, len(userConnections))
		for _, c := range userConnections {
			conns = append(conns, c)
		}
		bulletsToRemove = []string{}
		killsList = []*playerHitData{}
		newBullets = []*bulletData{}
		playersToSend = make(map[string]*playerData)
		mu.Unlock()

		for _, c := range conns {
			_ = c.writeJSON(payload)
		}
	}
}
