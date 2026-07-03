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
)

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

var BackgroundCards = make(map[int]map[int][]any)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func registerWebSocket(router *gin.Engine) {
	// TODO change to /ws after 1.0
	router.GET("/", func(c *gin.Context) {
		wsHandler(c.Writer, c.Request)
	})
	go broadCastInterval()
	playersToSend = make(map[string]*wsEvent)
}

type wsEvent struct {
	EventName string     `json:"eventName" bson:"eventName"`
	SocketId  string     `json:"socketId" bson:"socketId"`
	Data      [25][2]int `json:"data" bson:"data"`
	Bullet    bullet     `json:"bullet" bson:"bullet"`
	X         float32    `json:"x" bson:"x"` // TODO REMOVE used in newBullet event
	Y         float32    `json:"y" bson:"y"` // TODO REMOVE used in newBullet event

	// TODO
	//hitData    playerHitData `json:"playerHitData" bson:"playerHitData"`
	//playerData playerData    `json:"playerData" bson:"playerData"`

	// TODO move to subclass playerHit
	BulletId     string  `json:"bulletId" bson:"bulletId"`         // TODO move to subclass playerHit
	PlayerId     string  `json:"playerId" bson:"playerId"`         // TODO move to subclass playerHit
	From         string  `json:"from" bson:"from"`                 // TODO move to subclass playerHit
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"` // TODO move to subclass playerHit
	// TODO move to subclass playerHit

	// TODO moveToSubClass player
	// x, y, socketId?, eventName?
	Credits      int     `json:"credits" bson:"credits"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	Scale        float32 `json:"scale" bson:"scale"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
	Name         string  `json:"name" bson:"name"`
	Life         float32 `json:"life" bson:"life"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`

	// TODO moveToSubClass player
}

/*
type PlayerData struct {
	X            int     `json:"x" bson:"x"`
	Y            int     `json:"y" bson:"y"`
	Credits      int     `json:"credits" bson:"credits"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	Scale        float32 `json:"scale" bson:"scale"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
	Name         string  `json:"name" bson:"name"`
	Life         float32 `json:"life" bson:"life"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`
}

type playerHitData struct {
	BulletId     string  `json:"bulletId" bson:"bulletId"`
	PlayerId     string  `json:"playerId" bson:"playerId"`
	From         string  `json:"from" bson:"from"`
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"`
}*/

type bullet struct {
	Angle         float32 `json:"angle" bson:"angle"`
	BulletCharge  float32 `json:"bulletCharge" bson:"bulletCharge"`
	ExpY          float32 `json:"expY" bson:"expY"`
	ExpX          float32 `json:"expX" bson:"expX"`
	Id            string  `json:"id" bson:"id"`
	MoveX         float32 `json:"moveX" bson:"moveX"`
	MoveY         float32 `json:"moveY" bson:"moveY"`
	Rotation      float32 `json:"rotation" bson:"rotation"`
	ShootingSpeed float32 `json:"shootingSpeed" bson:"shootingSpeed"`
	X             float32 `json:"x" bson:"x"`
	Y             float32 `json:"y" bson:"y"`
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
		// TODO FIX THIS RETURNING MESSAGE AND DO A V2 OF THIS MESSAGE AFTER MIGRATION
		// TODO This is done in this way for retrocompatibility but it's bullshit
		wsGetBackgroundCards(conn, msg)
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

func wsGetBackgroundCards(conn *websocket.Conn, wsBgCards *wsEvent) {
	result := []any{}
	for card := range wsBgCards.Data {
		var x = wsBgCards.Data[card][0]
		var y = wsBgCards.Data[card][1]
		bgcX, xOk := BackgroundCards[x]
		if !xOk {
			bgcX = make(map[int][]any)
			BackgroundCards[x] = bgcX
		}
		_, yOk := BackgroundCards[x][y]
		if !yOk {

			points := make([][3]int, 0, 500)
			rw := resolutions[currentResolution].Width
			rh := resolutions[currentResolution].Height

			for i := 0; i < 500; i++ {
				point := [3]int{
					rand.IntN(rw),
					rand.IntN(rh),
					rand.IntN(4) + 1,
				}
				points = append(points, point)
			}
			BackgroundCards[x][y] = []any{x, y, points}
		}
		result = append(result, BackgroundCards[x][y])
		//log.Println(fmt.Sprintf("Card %d: x=%d, y=%d, xInCard=%d, yInCard=%d, size=%d", card, x, y, bgcY[0], bgcY[1], bgcY[2]))
	}
	_ = conn.WriteJSON(gin.H{"eventName": wsBgCards.EventName, "socketId": wsBgCards.SocketId, "cards": result})
}

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	for range ticker.C {
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
