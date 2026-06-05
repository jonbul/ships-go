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
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
}
type wsEventBackgroundCards struct {
	EventName string     `json:"eventName" bson:"eventName"`
	SocketId  string     `json:"socketId" bson:"socketId"`
	Data      [25][2]int `json:"data" bson:"data"`
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
		fmt.Println("Received: %s\\n", msg.EventName)

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
		var messageJson map[string]any
		_ = json.Unmarshal(messageByte, &messageJson)
		if nil != messageJson {
			playersToSend = append(playersToSend, messageJson)
			if nil != messageJson["socketId"] {
				players[messageJson["socketId"].(string)] = messageJson
			} else {
				log.Println("Error parsing playersToSend -> " + string(messageByte))
			}
		}
		return
	case "getBackgroundCards":
		// TODO FIX THIS RETURNING MESSAGE AND DO A V2 OF THIS MESSAGE AFTER MIGRATION
		// TODO This is done in this way for retrocompatibility but it's bullshit
		var wsBgCards = wsEventBackgroundCards{}
		_ = json.Unmarshal(messageByte, &wsBgCards)
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
		_ = conn.WriteJSON(gin.H{"eventName": msg.EventName, "socketId": msg.SocketId, "cards": result})
		return
	case "newBullet":
		// TODO
		log.Println("--------------------------")
		log.Println("Implementation pending: " + msg.EventName)
		log.Println("--------------------------")
		return
	case "playerHit":
		// TODO
		log.Println("--------------------------")
		log.Println("Implementation pending: " + msg.EventName)
		log.Println("--------------------------")
		return
	default:
		log.Println("--------------------------")
		log.Println("Unknown event: " + msg.EventName)
		log.Println("--------------------------")
	}
}

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	for range ticker.C {
		//log.Println(fmt.Sprintf("Sending event: %s", eventName))
		if len(playersToSend) == 0 {
			continue
		}

		var json = map[string]any{
			"eventName":       "gameBroadcast",
			"bulletsToRemove": []any{}, // TODO
			"newBullets":      []any{}, // TODO
			"players":         playersToSend,
			"kills":           []any{}, // TODO
		}

		for _, c := range userConnections {
			_ = c.WriteJSON(json)
		}
		playersToSend = []any{}
		defer ticker.Stop()
	}
}
