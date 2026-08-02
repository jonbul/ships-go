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

type bullet = models.Bullet
type wsEvent = models.WsEvent

var mu sync.Mutex
var userConnections = make(map[string]*websocket.Conn)
var playerStatus = make(map[string]*websocket.Conn)
var playersToSend = make(map[string]*wsEvent)
var players = make(map[string]*wsEvent)
var newBullets = []bullet{}
var bulletsToRemove = []string{}
var killsList = []*wsEvent{}
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
	playersToSend = make(map[string]*wsEvent)
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
	mu.Lock()
	userConnections[socketId] = conn
	mu.Unlock()
	// Listen for incoming messages
	for {
		// Read message from the client
		var msg wsEvent
		_, messagePlain, err := conn.ReadMessage()
		_ = json.Unmarshal(messagePlain, &msg)
		if err != nil {
			break
		}

		manageInputMessage(conn, &msg, socketId)
	}
}

func manageInputMessage(conn *websocket.Conn, msg *wsEvent, socketId string) {
	switch msg.EventName {
	case "connectionSuccess":
		log.Println("New connection with socketId: " + socketId)
		mu.Lock()
		userConnections[socketId] = conn
		mu.Unlock()
		msg.SocketId = socketId
		_ = conn.WriteJSON(msg)
		return
	case "playerData":
		if "" != socketId && "" != msg.SocketId {
			msg.SocketId = socketId
			mu.Lock()
			playersToSend[socketId] = msg
			players[socketId] = msg
			mu.Unlock()
		}
		return
	case "getBackgroundCards":
		wsGetBackgroundCards(conn)
		return
	case "newBullet":
		mu.Lock()
		newBullets = append(newBullets, msg.Bullet)
		mu.Unlock()
		return
	case "removeBullet":
		mu.Lock()
		bulletsToRemove = append(bulletsToRemove, msg.BulletId)
		mu.Unlock()
		return
	case "playerHit":
		mu.Lock()
		targetConn := userConnections[msg.PlayerId]
		mu.Unlock()
		if targetConn != nil {
			_ = targetConn.WriteJSON(msg)
		}
		return
	case "playerDied":
		mu.Lock()
		killsList = append(killsList, msg)
		playerFrom, ok := players[msg.From]
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
	return
}

func wsGetBackgroundCards(conn *websocket.Conn) {
	buildBackgroundCards()
	_ = conn.WriteJSON(gin.H{"eventName": "getBackgroundCards", "cards": BackgroundCards})
}

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	for range ticker.C {
		ActivePlayers.Set(float64(len(players)))
		mu.Lock()
		if len(playersToSend)+len(newBullets)+len(killsList) == 0 {
			mu.Unlock()
			continue
		}

		var payload = map[string]any{
			"eventName":       "gameBroadcast",
			"bulletsToRemove": bulletsToRemove,
			"newBullets":      newBullets,
			"players":         playersToSend,
			"kills":           killsList,
		}

		conns := make([]*websocket.Conn, 0, len(userConnections))
		for _, c := range userConnections {
			conns = append(conns, c)
		}
		bulletsToRemove = []string{}
		killsList = []*wsEvent{}
		newBullets = []bullet{}
		playersToSend = make(map[string]*wsEvent)
		mu.Unlock()

		for _, c := range conns {
			_ = c.WriteJSON(payload)
		}
	}
}
