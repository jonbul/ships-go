package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"

	"ships/controllers/websocket/models"
)

type bulletData = models.BulletData
type wsEvent = models.WsEvent
type playerData = models.PlayerData
type playerHitData = models.PlayerHitData
type npcData = models.NpcData

type safeConn struct {
	conn       *websocket.Conn
	mu         sync.Mutex
	remoteAddr string
	isNpc      bool
}

func (sc *safeConn) writeJSON(v any) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.conn.WriteJSON(v)
}

var mu sync.Mutex
var userConnections = make(map[string]*safeConn)
var playersToSend = make(map[string]*playerData)
var Players = make(map[string]*playerData)
var newBullets = []*bulletData{}
var bulletsToRemove = []string{}
var killsList = []*playerHitData{}
var hasPlayersTosend = false

// npcs holds the latest full snapshot of every NPC (black holes and future
// NPC kinds), as pushed by the ships-npc service. ships-go doesn't simulate
// NPCs itself anymore: it just relays this snapshot to players.
var npcs = make(map[string]npcData)

var cardSizeX = 3840
var cardSizeY = 3840

var ActivePlayers = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ships_active_players",
	Help: "Current number of in-game players.",
})

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

func RegisterWebSocket(router *gin.Engine) {
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
		delete(Players, socketId)
		mu.Unlock()
	}(conn, socketId)
	sc := &safeConn{conn: conn, remoteAddr: r.RemoteAddr}
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
	var rawMsgs []json.RawMessage
	if err := json.Unmarshal(msgPlain, &rawMsgs); err != nil {
		log.Println("invalid ws payload:", err)
		return
	}

	for _, raw := range rawMsgs {
		var meta struct {
			EventName string `json:"eventName"`
			SocketId  string `json:"socketId"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			log.Println("invalid ws item:", err)
			continue
		}
		var msg wsEvent
		_ = json.Unmarshal(raw, &msg)
		switch msg.EventName {
		case "connectionSuccess":
			log.Println("New connection with socketId: " + socketId)
			msg.SocketId = socketId
			_ = conn.writeJSON(msg)
		case "playerData":

			var plData *playerData
			_ = json.Unmarshal(raw, &plData)
			if "" != socketId && "" != msg.SocketId && "" != plData.SocketId {
				msg.SocketId = socketId
				mu.Lock()
				playersToSend[socketId] = plData
				Players[socketId] = plData
				mu.Unlock()
			}
		case "getBackgroundCards":
			wsGetBackgroundCards(conn)
		case "newBullet":

			var bullet *bulletData
			_ = json.Unmarshal(raw, &bullet)
			mu.Lock()
			newBullets = append(newBullets, bullet)
			mu.Unlock()
		case "removeBullet":

			var playerHit *playerHitData
			_ = json.Unmarshal(raw, &playerHit)
			mu.Lock()
			bulletsToRemove = append(bulletsToRemove, playerHit.BulletId)
			mu.Unlock()
		case "playerHit":
			var playerHit *playerHitData
			_ = json.Unmarshal(raw, &playerHit)
			mu.Lock()
			targetConn := userConnections[playerHit.PlayerId]
			mu.Unlock()
			if targetConn != nil {
				_ = targetConn.writeJSON(playerHit)
			}
		case "playerDied":
			var plHitData *playerHitData
			_ = json.Unmarshal(raw, &plHitData)
			mu.Lock()
			killsList = append(killsList, plHitData)
			playerFrom, ok := Players[plHitData.From]
			if ok {
				hasPlayersTosend = true
				playerFrom.Credits += 100
			}
			mu.Unlock()
		case "npcAuth":
			var auth models.NpcAuthData
			_ = json.Unmarshal(raw, &auth)
			if isNpcAuthValid(conn, auth.Secret) {
				conn.mu.Lock()
				conn.isNpc = true
				conn.mu.Unlock()
				log.Println("NPC controller authenticated from " + conn.remoteAddr)
			} else {
				log.Println("Rejected npcAuth attempt from " + conn.remoteAddr)
			}
		case "npcUpdate":
			if !conn.isNpc {
				log.Println("Ignoring npcUpdate from unauthenticated connection " + conn.remoteAddr)
				continue
			}
			var update models.NpcUpdateData
			_ = json.Unmarshal(raw, &update)
			mu.Lock()
			npcs = make(map[string]npcData, len(update.Npcs))
			for _, npc := range update.Npcs {
				npcs[npc.Id] = npc
			}
			mu.Unlock()
		default:
			log.Println("--------------------------")
			log.Println("Unknown event: " + msg.EventName)
			log.Println("--------------------------")
		}
	}
}

func buildBackgroundCards() {
	if len(BackgroundCards) > 0 {
		return
	}
	var w = cardSizeX
	var h = cardSizeY

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
	_ = conn.writeJSON(gin.H{"eventName": "getBackgroundCards", "cards": BackgroundCards, "cardSize": gin.H{"x": cardSizeX, "y": cardSizeY}})
}

var lastBroadcastTime int64 = 0

func broadCastInterval() {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()
	for range ticker.C {
		broadCastIntervalLoop()
	}
}

func broadCastIntervalLoop() {
	ActivePlayers.Set(float64(len(Players)))
	mu.Lock()
	defer mu.Unlock()

	var currentTime = time.Now().UnixMilli()

	// send at least every 2 seconds or if there are any new bullets, players, or kills to send
	if len(Players) == 0 || (len(playersToSend)+len(newBullets)+len(killsList)+len(npcs) == 0 && currentTime-lastBroadcastTime < 2000) {
		return
	}
	lastBroadcastTime = currentTime

	var playerIds = make([]string, 0, len(Players))
	for id := range Players {
		playerIds = append(playerIds, id)
	}

	var payload = map[string]any{
		"eventName":       "gameBroadcast",
		"bulletsToRemove": bulletsToRemove,
		"newBullets":      newBullets,
		"players":         playersToSend,
		"kills":           killsList,
		"activePlayerIds": playerIds,
		"npcs":            npcs,
	}

	conns := make([]*safeConn, 0, len(userConnections))
	for _, c := range userConnections {
		conns = append(conns, c)
	}
	bulletsToRemove = []string{}
	killsList = []*playerHitData{}
	newBullets = []*bulletData{}
	playersToSend = make(map[string]*playerData)

	for _, c := range conns {
		_ = c.writeJSON(payload)
	}
}

// isNpcAuthValid only allows NPC controller connections (ships-npc) coming
// from localhost and presenting the shared secret configured via the
// NPC_SECRET env var, so a stray websocket client can't inject fake NPCs.
func isNpcAuthValid(conn *safeConn, secret string) bool {
	expectedSecret := os.Getenv("NPC_SECRET")
	if expectedSecret == "" || secret != expectedSecret {
		return false
	}
	return isLoopbackAddr(conn.remoteAddr)
}

func isLoopbackAddr(remoteAddr string) bool {
	host := remoteAddr
	if h, _, err := net.SplitHostPort(remoteAddr); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
