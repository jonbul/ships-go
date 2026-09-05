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

// NpcData is the generic shape used for every NPC kind (black hole, and any
// future NPC types). NPCs are simulated by the ships-npc service and pushed
// to ships-go over a websocket connection, similarly to how a player pushes
// playerData.
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

// private struct to hold NPC types
type npcTypes struct {
	BlackHole string `json:"blackHole" bson:"blackHole"`
}

// public instance to use as enum
var NpcTypes npcTypes = npcTypes{
	BlackHole: "BlackHole",
}
