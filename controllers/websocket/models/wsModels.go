package models

type WsEvent struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
}

type PlayerData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	Credits      int     `json:"credits" bson:"credits"`
	Deaths       int     `json:"deaths" bson:"deaths"`
	Hide         bool    `json:"hidden" bson:"hidden"`
	IsDead       bool    `json:"isDead" bson:"isDead"`
	Kills        int     `json:"kills" bson:"kills"`
	Life         float32 `json:"life" bson:"life"`
	Name         string  `json:"name" bson:"name"`
	Rotate       float32 `json:"rotate" bson:"rotate"`
	Scale        float32 `json:"scale" bson:"scale"`
	ShipId       string  `json:"shipId" bson:"shipId"`
	X            float32 `json:"x" bson:"x"`
	Xtranslation float32 `json:"xTranslation" bson:"xTranslation"`
	Y            float32 `json:"y" bson:"y"`
	YTranslation float32 `json:"yTranslation" bson:"yTranslation"`
}

type PlayerHitData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"`
	BulletId     string  `json:"bulletId" bson:"bulletId"`
	From         string  `json:"from" bson:"from"`
	PlayerId     string  `json:"playerId" bson:"playerId"`
	X            float32 `json:"x" bson:"x"`
	Y            float32 `json:"y" bson:"y"`
}

type BulletData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
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

// NpcData is the generic shape used for every NPC kind (black hole, enemy
// ship, and any future NPC types). NPCs are simulated by the ships-npc
// service and pushed to ships-go over a websocket connection, similarly to
// how a player pushes playerData. Fields only meaningful for one kind (e.g.
// ShipId/Life for Ship NPCs) are simply left zero-valued for the others.
type NpcData struct {
	Type      string  `json:"type" bson:"type"`
	Id        string  `json:"id" bson:"id"`
	X         float64 `json:"x" bson:"x"`
	Y         float64 `json:"y" bson:"y"`
	Scale     float64 `json:"scale" bson:"scale"`
	MaxSize   int     `json:"maxSize" bson:"maxSize"`
	Direction float64 `json:"direction" bson:"direction"`
	Duration  int     `json:"duration" bson:"duration"`
	Speed     float64 `json:"speed" bson:"speed"`
	// Ship NPC fields (Type == NpcTypes.Ship): a hostile NPC using one of
	// the public ships, rendered/collided exactly like a player.
	ShipId  string  `json:"shipId,omitempty" bson:"shipId,omitempty"`
	Name    string  `json:"name,omitempty" bson:"name,omitempty"`
	Rotate  float32 `json:"rotate,omitempty" bson:"rotate,omitempty"`
	Life    float32 `json:"life,omitempty" bson:"life,omitempty"`
	MaxLife float32 `json:"maxLife,omitempty" bson:"maxLife,omitempty"`
}

// NpcAuthData is sent once by ships-npc right after connecting, to
// authenticate the connection as an NPC controller rather than a player.
type NpcAuthData struct {
	EventName string `json:"eventName" bson:"eventName"`
	Secret    string `json:"secret" bson:"secret"`
}

// NpcUpdateData carries the full, current set of NPCs simulated by
// ships-npc. Sending the whole batch each time (instead of per-NPC deltas)
// keeps ships-go stateless/authoritative-follower and avoids desync, while
// still being cheap since the NPC count is small.
type NpcUpdateData struct {
	EventName string    `json:"eventName" bson:"eventName"`
	Npcs      []NpcData `json:"npcs" bson:"npcs"`
}

// NpcSettingsData is the set of NPC knobs an administrator can change at
// runtime from ships-vue's admin panel. ships-go holds the current values
// (the browser can only reach ships-go, never ships-npc directly) and
// pushes them to the NPC controller connection as an NpcConfigData event.
// ships-npc mirrors this struct and is the one that clamps/validates them.
type NpcSettingsData struct {
	EnemyShips              int     `json:"enemyShips" bson:"enemyShips"`
	EnemyShipLife           float32 `json:"enemyShipLife" bson:"enemyShipLife"`
	EnemyShipSpeed          float64 `json:"enemyShipSpeed" bson:"enemyShipSpeed"`
	EnemyShipFireRateMs     int     `json:"enemyShipFireRateMs" bson:"enemyShipFireRateMs"`
	MaxBlackHoles           int     `json:"maxBlackHoles" bson:"maxBlackHoles"`
	BlackHoleSpawnPeriodSec int     `json:"blackHoleSpawnPeriodSec" bson:"blackHoleSpawnPeriodSec"`
}

// Sanitized clamps settings into workable ranges. ships-npc clamps again
// on receipt (it must, since it can't trust the wire), but doing it here
// too means the admin panel can echo back the values that are genuinely in
// force instead of showing a number that was silently rejected downstream.
// The bounds are duplicated in ships-npc's npcSettings.sanitized().
func (s NpcSettingsData) Sanitized() NpcSettingsData {
	s.EnemyShips = clampInt(s.EnemyShips, 0, 20)
	if s.EnemyShipLife <= 0 {
		s.EnemyShipLife = 10
	}
	if s.EnemyShipSpeed <= 0 {
		s.EnemyShipSpeed = 20
	}
	if s.EnemyShipSpeed > maxGameSpeed {
		s.EnemyShipSpeed = maxGameSpeed
	}
	s.EnemyShipFireRateMs = clampInt(s.EnemyShipFireRateMs, 100, 600000)
	s.MaxBlackHoles = clampInt(s.MaxBlackHoles, 0, 50)
	s.BlackHoleSpawnPeriodSec = clampInt(s.BlackHoleSpawnPeriodSec, 1, 86400)
	return s
}

// maxGameSpeed mirrors ships-vue's SPEED.MAX: enemy ship speed is
// expressed in the game's own speed units, so this is "as fast as a player
// at full throttle".
const maxGameSpeed = 50

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

// NpcConfigData is the envelope carrying NpcSettingsData to ships-npc.
type NpcConfigData struct {
	EventName string          `json:"eventName" bson:"eventName"`
	Settings  NpcSettingsData `json:"settings" bson:"settings"`
}

// NpcHitData is sent by a player's client when one of its own bullets hits
// a Ship NPC (see checkBulletCollision client-side). ships-go doesn't track
// NPC health itself: it just forwards this to the NPC controller
// connection (ships-npc), which is the source of truth for NPC state.
type NpcHitData struct {
	EventName    string  `json:"eventName" bson:"eventName"`
	NpcId        string  `json:"npcId" bson:"npcId"`
	BulletId     string  `json:"bulletId" bson:"bulletId"`
	From         string  `json:"from" bson:"from"`
	BulletCharge float32 `json:"bulletCharge" bson:"bulletCharge"`
	X            float32 `json:"x" bson:"x"`
	Y            float32 `json:"y" bson:"y"`
}

// private struct to hold NPC types
type npcTypes struct {
	BlackHole string `json:"blackHole" bson:"blackHole"`
	Ship      string `json:"ship" bson:"ship"`
}

// public instance to use as enum
var NpcTypes npcTypes = npcTypes{
	BlackHole: "BlackHole",
	Ship:      "Ship",
}
