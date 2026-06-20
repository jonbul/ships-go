package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var userConnections = make(map[string]*websocket.Conn)
var playerStatus = make(map[string]*websocket.Conn)
var playersToSend = []any{}
var players = gin.H{}
var newBullets = []bullet{}
var bulletsToRemove = []string{}
var killsList = []wsEvent{}
var hasPlayersTosend = false

// backgroundCards
// {x: {y : [xInCard, yInCard,size(1 to 5)]}}
// {1: {1 : [1,2,3],2 : [1,2,3]}, 2: {1 : [1,2,3],2 : [1,2,3]...}...}

var backgroundCards = make(map[int]map[int][]any)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func RegisterWebSocket(router *gin.Engine) {
	// TODO change to /ws after 1.0
	router.GET("/", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	go broadCastInterval()
	playersToSend = []any{}
}

type wsEvent struct {
	EventName string     `json:"eventName" bson:"eventName"`
	SocketId  string     `json:"socketId" bson:"socketId"`
	Data      [25][2]int `json:"data" bson:"data"`
	Bullet    bullet     `json:"bullet" bson:"bullet"`
	X         int        `json:"x" bson:"x"` // TODO REMOVE used in newBullet event
	Y         int        `json:"y" bson:"y"` // TODO REMOVE used in newBullet event
	// TODO move to subclass playerHit
	BulletId     string  `json:"bulletId" bson:"bulletId"`         // TODO move to subclass playerHit
	PlayerId     string  `json:"playerId" bson:"playerId"`         // TODO move to subclass playerHit
	From         string  `json:"from" bson:"from"`                 // TODO move to subclass playerHit
	BulletCharge float64 `json:"bulletCharge" bson:"bulletCharge"` // TODO move to subclass playerHit
	// TODO move to subclass playerHit
}

type bullet struct {
	Angle         float64 `json:"angle" bson:"angle"`
	BulletCharge  float64 `json:"bulletCharge" bson:"bulletCharge"`
	ExpY          float64 `json:"expY" bson:"expY"`
	ExpX          float64 `json:"expX" bson:"expX"`
	Id            string  `json:"id" bson:"id"`
	MoveX         float64 `json:"moveX" bson:"moveX"`
	MoveY         float64 `json:"moveY" bson:"moveY"`
	Rotation      float64 `json:"rotation" bson:"rotation"`
	ShootingSpeed float64 `json:"shootingSpeed" bson:"shootingSpeed"`
	X             float64 `json:"x" bson:"x"`
	Y             float64 `json:"y" bson:"y"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection:", err)
		}
	}(conn)
	//defer conn.Close()
	// Listen for incoming messages
	for {
		// Read message from the client
		var msg wsEvent
		_, messagePlain, err := conn.ReadMessage()
		_ = json.Unmarshal(messagePlain, &msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received: %s\n", msg.EventName)

		manageInputMessage(conn, &msg, messagePlain)
	}
}

func manageInputMessage(conn *websocket.Conn, msg *wsEvent, messageByte []byte) {
	switch msg.EventName {
	case "connectionSuccess":
		socketId := uuid.New().String()
		log.Println("New connection with socketId: " + socketId)
		userConnections[socketId] = conn
		_ = conn.WriteJSON(wsEvent{EventName: "connectionSuccess", SocketId: socketId})
		return
	case "playerData":
		// string to generic json
		var playerEvent map[string]any
		_ = json.Unmarshal(messageByte, &playerEvent)
		if nil != playerEvent {
			playersToSend = append(playersToSend, playerEvent)
			if nil != playerEvent["socketId"] {
				players[playerEvent["socketId"].(string)] = playerEvent
			} else {
				log.Println("Error parsing playersToSend -> " + string(messageByte))
			}
		}
		return
	case "getBackgroundCards":
		// TODO FIX THIS RETURNING MESSAGE AND DO A V2 OF THIS MESSAGE AFTER MIGRATION
		// TODO This is done in this way for retrocompatibility but it's bullshit
		wsGetBackgroundCards(conn, msg)
		return
	case "newBullet":
		newBullets = append(newBullets, msg.Bullet)
		return
	case "removeBullet":
		bulletsToRemove = append(bulletsToRemove, msg.BulletId)
		return
	case "playerHit":
		userConnections[msg.PlayerId].WriteJSON(msg)
		return
	case "playerDied":
		killsList = append(killsList, *msg)
		if players[msg.From] != nil {
			hasPlayersTosend = true
		}
	default:
		log.Println("--------------------------")
		log.Println("Unknown event: " + msg.EventName)
		log.Println("--------------------------")
	}
}

func wsGetBackgroundCards(conn *websocket.Conn, wsBgCards *wsEvent) {
	result := []any{}
	for card := range wsBgCards.Data {
		var x = wsBgCards.Data[card][0]
		var y = wsBgCards.Data[card][1]
		bgcX, xOk := backgroundCards[x]
		if !xOk {
			bgcX = make(map[int][]any)
			backgroundCards[x] = bgcX
		}
		_, yOk := backgroundCards[x][y]
		if !yOk {

			points := make([][3]int, 0, 500)
			for i := 0; i < 500; i++ {
				point := [3]int{
					rand.IntN(canvasWidth),
					rand.IntN(canvasHeight),
					rand.IntN(4) + 1,
				}
				points = append(points, point)
			}
			backgroundCards[x][y] = []any{x, y, points}
		}
		result = append(result, backgroundCards[x][y])
		//log.Println(fmt.Sprintf("Card %d: x=%d, y=%d, xInCard=%d, yInCard=%d, size=%d", card, x, y, bgcY[0], bgcY[1], bgcY[2]))
	}
	_ = conn.WriteJSON(gin.H{"eventName": wsBgCards.EventName, "socketId": wsBgCards.SocketId, "cards": result})
}

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	for range ticker.C {
		//log.Println(fmt.Sprintf("Sending event: %s", eventName))
		if len(playersToSend)+len(newBullets)+len(killsList) == 0 {
			continue
		}

		var json = map[string]any{
			"eventName":       "gameBroadcast",
			"bulletsToRemove": bulletsToRemove, // TODO
			"newBullets":      newBullets,      // TODO
			"players":         playersToSend,
			"kills":           killsList, // TODO
		}

		for _, c := range userConnections {
			_ = c.WriteJSON(json)
		}
		bulletsToRemove = []string{}
		killsList = []wsEvent{}
		newBullets = []bullet{}
		playersToSend = []any{}
		defer ticker.Stop()
	}
}
