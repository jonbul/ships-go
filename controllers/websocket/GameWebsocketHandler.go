package websocket

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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
type npcHitData = models.NpcHitData

type safeConn struct {
	conn       *websocket.Conn
	mu         sync.Mutex
	remoteAddr string
	// isNpc is atomic because it is written by this connection's own reader
	// goroutine (on npcAuth) but read from others: any player's reader
	// goroutine relaying an npcHit, and the admin HTTP goroutine pushing
	// new settings. Guarding it with conn.mu while those readers hold the
	// global mu instead would be no synchronization at all.
	isNpc atomic.Bool
}

// wsWriteTimeout bounds a single frame write. Without it a client that
// simply stops reading fills its TCP send buffer and the write blocks
// forever - and callers here write while holding the global game mutex, so
// one such connection would stall the game for everybody.
const wsWriteTimeout = 5 * time.Second

func (sc *safeConn) writeJSON(v any) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	_ = sc.conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	err := sc.conn.WriteJSON(v)
	if err != nil {
		// A failed write leaves a websocket unusable. Closing it makes the
		// reader goroutine return so the normal disconnect cleanup runs,
		// instead of leaving a dead socket in userConnections forever.
		_ = sc.conn.Close()
	}
	return err
}

var mu sync.Mutex
var userConnections = make(map[string]*safeConn)
var playersToSend = make(map[string]*playerData)
var Players = make(map[string]*playerData)
var newBullets = []*bulletData{}
var bulletsToRemove = []string{}
var killsList = []*playerHitData{}

// npcs holds the latest full snapshot of every NPC (black holes and future
// NPC kinds), as pushed by the ships-npc service. ships-go doesn't simulate
// NPCs itself anymore: it just relays this snapshot to players.
var npcs = make(map[string]npcData)

// npcControllerId is the socketId of the one connection currently allowed to
// drive the NPCs. Every npcUpdate replaces the whole snapshot, so two
// controllers (a stray instance left running next to a restarted one) do not
// add up: they overwrite each other every tick. Players then see NPCs flicker
// in and out of existence - unnamed, undrawn and impossible to hit, while
// their bullets keep arriving. The first controller to authenticate keeps the
// role until it disconnects, so a duplicate is refused instead of corrupting
// the game.
var npcControllerId string

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
		// Releasing the NPC controller role also drops its NPCs: they are
		// only alive as long as something is simulating them, and leaving
		// the last snapshot behind would strand ghost ships that never move
		// and can never be killed.
		if socketId == npcControllerId {
			npcControllerId = ""
			npcs = make(map[string]npcData)
			log.Println("NPC controller disconnected, NPCs cleared")
		}
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
			// Sent straight after, so a client that joins mid-game starts
			// out enforcing the rules currently in force rather than the
			// defaults compiled into it.
			sendGameSettings(conn)
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
				playerFrom.Credits += 100
				// Queue the killer so the new credit total actually reaches
				// the clients: without this it only ships out on that
				// player's next playerData frame. (This replaces a
				// `hasPlayersTosend = true` flag that nothing ever read.)
				playersToSend[plHitData.From] = playerFrom
			}
			mu.Unlock()
		case "npcAuth":
			var auth models.NpcAuthData
			_ = json.Unmarshal(raw, &auth)
			if !isNpcAuthValid(conn, auth.Secret) {
				log.Println("Rejected npcAuth attempt from " + conn.remoteAddr)
				break
			}

			mu.Lock()
			_, incumbentAlive := userConnections[npcControllerId]
			taken := npcControllerId != "" && npcControllerId != socketId && incumbentAlive
			if !taken {
				npcControllerId = socketId
			}
			mu.Unlock()

			if taken {
				// Refusing is safer than taking over: two live controllers
				// would otherwise evict each other in a loop. A genuine
				// restart frees the role as soon as its old socket closes.
				log.Println("Refusing second NPC controller from " + conn.remoteAddr +
					": one is already connected")
				_ = conn.writeJSON(gin.H{
					"eventName": "npcRejected",
					"reason":    "another NPC controller is already connected",
				})
				break
			}

			conn.isNpc.Store(true)
			log.Println("NPC controller authenticated from " + conn.remoteAddr)
			// Tell it the settings in force right away, so a restart
			// of either process converges without an admin having to
			// re-save the panel.
			sendNpcSettings(conn)
		case "npcUpdate":
			if !conn.isNpc.Load() {
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
		case "npcHit":
			// A player's client detected that its own bullet hit a Ship NPC
			// (see checkBulletCollision in ships-vue). ships-go doesn't
			// track NPC health itself, so just forward this to whichever
			// connection(s) are the authenticated NPC controller
			// (ships-npc), which owns that state.
			var hit npcHitData
			_ = json.Unmarshal(raw, &hit)
			// Collect under the lock, write after releasing it: a slow or
			// dead socket must never block the game loop (the same rule
			// SetNpcSettings follows).
			mu.Lock()
			npcConns := make([]*safeConn, 0, 1)
			for _, c := range userConnections {
				if c.isNpc.Load() {
					npcConns = append(npcConns, c)
				}
			}
			mu.Unlock()
			for _, c := range npcConns {
				_ = c.writeJSON(hit)
			}
		default:
			log.Println("--------------------------")
			log.Println("Unknown event: " + msg.EventName)
			log.Println("--------------------------")
		}
	}
}

// backgroundCardsOnce guards the lazy build below. Every client calls
// getBackgroundCards on join from its own reader goroutine, so a plain
// `if len(...) > 0 { return }` guard both races the map writes (a fatal,
// unrecoverable "concurrent map writes") and lets a second caller observe a
// half-built map. Once() makes the build happen exactly once and blocks
// later callers until it is finished; the map is read-only afterwards.
var backgroundCardsOnce sync.Once

func buildBackgroundCards() {
	backgroundCardsOnce.Do(buildBackgroundCardsOnce)
}

func buildBackgroundCardsOnce() {
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
	payload, conns := collectBroadcast()
	if payload == nil {
		return
	}
	// Written after the lock is released: a slow or dead socket must never
	// stall the game loop for everybody else. writeJSON has its own
	// deadline and closes the connection on failure.
	for _, c := range conns {
		_ = c.writeJSON(payload)
	}
}

// collectBroadcast builds this tick's payload and the list of connections to
// send it to, draining the per-tick buffers. It returns a nil payload when
// there is nothing to send. All shared state is touched here, under mu, and
// nowhere else in the broadcast path.
func collectBroadcast() (map[string]any, []*safeConn) {
	mu.Lock()
	defer mu.Unlock()

	// Read under the lock: Players is written by every connection's reader
	// goroutine, so even len() is a race from this ticker's goroutine.
	ActivePlayers.Set(float64(len(Players)))

	var currentTime = time.Now().UnixMilli()

	if len(Players) == 0 {
		// Nobody is playing, so there is nothing to render - but the
		// buffers must still be drained. ships-npc stays connected and
		// keeps firing, and anything left here would be retained until
		// somebody joined and then delivered as one enormous frame.
		drainBroadcastBuffers()
		// The NPC controller is still sent the (now empty) player list
		// every 2s: otherwise it never learns the last player left, and
		// carries on chasing and shooting a ghost forever.
		if currentTime-lastBroadcastTime < 2000 {
			return nil, nil
		}
	} else if len(playersToSend)+len(newBullets)+len(killsList)+len(npcs) == 0 &&
		currentTime-lastBroadcastTime < 2000 {
		// Send at least every 2 seconds, or as soon as there are new
		// bullets, players or kills to send.
		return nil, nil
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
	drainBroadcastBuffers()

	return payload, conns
}

// drainBroadcastBuffers resets the per-tick accumulators. The payload keeps
// the old slices/map, so replacing them here (rather than truncating) is
// what makes handing them to the marshaller safe.
func drainBroadcastBuffers() {
	bulletsToRemove = []string{}
	killsList = []*playerHitData{}
	newBullets = []*bulletData{}
	playersToSend = make(map[string]*playerData)
}

// isNpcAuthValid only allows NPC controller connections (ships-npc) coming
// from localhost and presenting the shared secret configured via the
// NPC_SECRET env var, so a stray websocket client can't inject fake NPCs.
func isNpcAuthValid(conn *safeConn, secret string) bool {
	expectedSecret := os.Getenv("NPC_SECRET")
	if expectedSecret == "" {
		return false
	}
	// Constant-time: `!=` returns as soon as two bytes differ, which leaks
	// how much of the secret was right. The loopback check below makes this
	// hard to exploit remotely, but anything co-located on the host - or a
	// local reverse proxy, where every RemoteAddr is 127.0.0.1 - leaves the
	// secret as the only real control.
	if subtle.ConstantTimeCompare([]byte(secret), []byte(expectedSecret)) != 1 {
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

// SnapshotPlayers returns a copy of the live player map, taken under the
// game lock. Callers outside this package (the admin HTTP handlers) run on
// their own goroutines, and ranging over the live map while a player's
// reader goroutine writes to it is not a race the runtime tolerates: it
// aborts the whole process with "concurrent map iteration and map write".
func SnapshotPlayers() map[string]*playerData {
	mu.Lock()
	defer mu.Unlock()

	out := make(map[string]*playerData, len(Players))
	for id, p := range Players {
		out[id] = p
	}
	return out
}
