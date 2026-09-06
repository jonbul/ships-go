package models

type WsEvent struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
}

type PlayerData struct {
	EventName string `json:"eventName" bson:"eventName"`
	SocketId  string `json:"socketId" bson:"socketId"`
	//—————————————————————————————————————————————————————————————————————
	Credits int     `json:"credits" bson:"credits"`
	Deaths  int     `json:"deaths" bson:"deaths"`
	Hide    bool    `json:"hide" bson:"hide"`
	IsDead  bool    `json:"isDead" bson:"isDead"`
	Kills   int     `json:"kills" bson:"kills"`
	Life    float32 `json:"life" bson:"life"`
	Name    string  `json:"name" bson:"name"`
	Rotate  float32 `json:"rotate" bson:"rotate"`
	Scale   float32 `json:"scale" bson:"scale"`
	ShipId  string  `json:"shipId" bson:"shipId"`
	// Width/Height are the player's raw (unscaled) ship size, sent by the
	// client because shipId alone cannot identify it: GET /game/getShips
	// only lists public ships, so a player's own painting project is a ship
	// nobody else can look up. ships-npc aims with these.
	Width        float32 `json:"width" bson:"width"`
	Height       float32 `json:"height" bson:"height"`
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
	ShipId string `json:"shipId,omitempty" bson:"shipId,omitempty"`
	Name   string `json:"name,omitempty" bson:"name,omitempty"`
	// Rotate is NOT omitempty: 0 is a perfectly legal heading (due east),
	// and it is exactly what a ship spawns with. Omitting it made
	// ships-vue's NPC update path assign `undefined`, which turned the
	// ship's position and collision box into NaN - an enemy ship that was
	// invisible and impossible to hit.
	Rotate  float32 `json:"rotate" bson:"rotate"`
	Life    float32 `json:"life,omitempty" bson:"life,omitempty"`
	MaxLife float32 `json:"maxLife,omitempty" bson:"maxLife,omitempty"`
	// Kills/Deaths let ships-vue list NPC ships in the scoreboard alongside
	// the players. Not omitempty: a zero score is meaningful and must still
	// reach the client instead of silently reading as "unknown".
	Kills  int `json:"kills" bson:"kills"`
	Deaths int `json:"deaths" bson:"deaths"`
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
	// EnemyShipController selects which brains fly the hostile ships:
	// "none", "rule" (the hand-written controller), "ai" (ships-npc's
	// embedded learned policy) or "both". Each fleet keeps its own size,
	// so switching away and back does not lose the number that was set.
	EnemyShipController string `json:"enemyShipController" bson:"enemyShipController"`
	EnemyShips          int    `json:"enemyShips" bson:"enemyShips"`
	AiShips             int    `json:"aiShips" bson:"aiShips"`
	// ShipLife is the life every ship enters the game with, players
	// included - the number ships-vue used to hardcode as 10. Like
	// ContactDamage it is a rule the browser enforces for itself, so it
	// travels in GameSettingsData too. Changing it does not heal or hurt
	// anybody already flying: a ship keeps the life it spawned with until
	// it dies, exactly as an NPC does.
	ShipLife                float32 `json:"shipLife" bson:"shipLife"`
	EnemyShipSpeed          float64 `json:"enemyShipSpeed" bson:"enemyShipSpeed"`
	EnemyShipFireRateMs     int     `json:"enemyShipFireRateMs" bson:"enemyShipFireRateMs"`
	MaxBlackHoles           int     `json:"maxBlackHoles" bson:"maxBlackHoles"`
	BlackHoleSpawnPeriodSec int     `json:"blackHoleSpawnPeriodSec" bson:"blackHoleSpawnPeriodSec"`
	// BlackHoleDurationSec is how long a black hole lives before it starts
	// shrinking away. Applied live, so shortening it can clear black holes
	// that are already on the map.
	BlackHoleDurationSec int `json:"blackHoleDurationSec" bson:"blackHoleDurationSec"`
	// ContactDamage is the odd one out: it applies to players as well as
	// NPCs, so it is also broadcast to every browser as a GameSettingsData
	// event. It lives here because it is one rule about one thing - what
	// happens when two ships touch - and splitting it across two settings
	// stores would only let the halves disagree.
	ContactDamage bool `json:"contactDamage" bson:"contactDamage"`
	// KillScaling makes a player's ship grow with (kills - deaths). Like
	// ContactDamage it is a rule the browser enforces for itself, so it
	// travels in GameSettingsData as well. Off by default: it is a large
	// change to how the game plays, and it also changes a ship's collision
	// box, so it is opt-in rather than something a new deployment inherits.
	KillScaling bool `json:"killScaling" bson:"killScaling"`
	// ShipSize is the size every ship is normalised to when it enters the
	// game, whatever its artwork measures - the number ships-vue used to
	// hardcode as 100. It is a rule about every ship, players included, so
	// it travels in GameSettingsData too, and ships-npc needs it to know
	// how big the ships it is flying actually are.
	ShipSize int `json:"shipSize" bson:"shipSize"`
	// The attack matrix: for each kind of attacker, which factions it may
	// hunt and damage. Any combination is allowed, including a fleet that
	// fights itself and one that attacks nothing.
	//
	// Deliberately none of them `omitempty`: false is a meaningful value
	// here, and an admin clearing a box must actually clear it downstream
	// rather than have the field vanish and leave ships-npc on its
	// previous value.
	NpcAttacksPlayers bool `json:"npcAttacksPlayers" bson:"npcAttacksPlayers"`
	NpcAttacksNpc     bool `json:"npcAttacksNpc" bson:"npcAttacksNpc"`
	NpcAttacksAi      bool `json:"npcAttacksAi" bson:"npcAttacksAi"`
	AiAttacksPlayers  bool `json:"aiAttacksPlayers" bson:"aiAttacksPlayers"`
	AiAttacksNpc      bool `json:"aiAttacksNpc" bson:"aiAttacksNpc"`
	AiAttacksAi       bool `json:"aiAttacksAi" bson:"aiAttacksAi"`
}

// GameSettingsData carries the rules a *player's* browser has to enforce
// itself. Ship-to-ship contact is resolved client-side (each client damages
// only itself, which comes out symmetric because every client runs the same
// check), so switching contact damage off has to reach the browsers as well
// as ships-npc, or players would keep hurting each other after an admin
// turned it off. Sent on connection and again whenever an admin saves.
type GameSettingsData struct {
	EventName     string  `json:"eventName"`
	ContactDamage bool    `json:"contactDamage"`
	KillScaling   bool    `json:"killScaling"`
	ShipSize      int     `json:"shipSize"`
	ShipLife      float32 `json:"shipLife"`
}

// Sanitized clamps settings into workable ranges. ships-npc clamps again
// on receipt (it must, since it can't trust the wire), but doing it here
// too means the admin panel can echo back the values that are genuinely in
// force instead of showing a number that was silently rejected downstream.
// The bounds are duplicated in ships-npc's npcSettings.sanitized().
func (s NpcSettingsData) Sanitized() NpcSettingsData {
	switch s.EnemyShipController {
	case NpcControllerNone, NpcControllerRule, NpcControllerAi, NpcControllerBoth:
	default:
		// An unrecognised value would otherwise mean "none" downstream,
		// so a typo or an older client would silently empty the map.
		s.EnemyShipController = NpcControllerRule
	}
	s.EnemyShips = clampInt(s.EnemyShips, 0, MaxNpcFleetSize)
	s.AiShips = clampInt(s.AiShips, 0, MaxNpcFleetSize)
	if s.ShipLife <= 0 {
		s.ShipLife = DefaultShipLife
	}
	if s.EnemyShipSpeed <= 0 {
		s.EnemyShipSpeed = 20
	}
	if s.EnemyShipSpeed > maxGameSpeed {
		s.EnemyShipSpeed = maxGameSpeed
	}
	s.EnemyShipFireRateMs = clampInt(s.EnemyShipFireRateMs, 100, 600000)
	// Zero means "not set" - an older settings document, or a client that
	// doesn't know about the field - and has to fall back to the default
	// rather than clamp up to the minimum, which would shrink every ship.
	if s.ShipSize <= 0 {
		s.ShipSize = DefaultShipSize
	}
	s.ShipSize = clampInt(s.ShipSize, MinShipSize, MaxShipSize)
	s.MaxBlackHoles = clampInt(s.MaxBlackHoles, 0, 50)
	s.BlackHoleSpawnPeriodSec = clampInt(s.BlackHoleSpawnPeriodSec, 1, 86400)
	// Zero means "not set" and takes the default, for the same reason as
	// ShipSize above: clamping up to the minimum would make every black
	// hole a blink-and-miss-it one.
	if s.BlackHoleDurationSec <= 0 {
		s.BlackHoleDurationSec = DefaultBlackHoleDurationSec
	}
	s.BlackHoleDurationSec = clampInt(s.BlackHoleDurationSec, 5, 86400)
	return s
}

// maxGameSpeed mirrors ships-vue's SPEED.MAX: enemy ship speed is
// expressed in the game's own speed units, so this is "as fast as a player
// at full throttle".
const maxGameSpeed = 50

// The controller choices, mirrored in ships-npc's settings.go.
const (
	NpcControllerNone = "none"
	NpcControllerRule = "rule"
	NpcControllerAi   = "ai"
	NpcControllerBoth = "both"
)

// MaxNpcFleetSize caps *each* fleet, so "both" at the maximum is twice
// this many ships. It is high enough to be a load test rather than a
// gameplay setting; ships-npc clamps to the same number.
const MaxNpcFleetSize = 100

// The bounds on the standard ship size, mirrored in ships-npc. The lower
// one keeps a ship big enough to see and to hit; the upper one keeps a
// fleet from filling the screen. DefaultShipSize is what ships-vue drew at
// before this was configurable, so an existing game plays identically.
// DefaultBlackHoleDurationSec is how long a black hole lived before this
// was configurable.
const DefaultBlackHoleDurationSec = 180

const (
	MinShipSize     = 20
	MaxShipSize     = 1000
	DefaultShipSize = 100
)

// DefaultShipLife is the life every ship - player, NPC or AI - used to be
// hardcoded with in ships-vue and ships-npc alike.
const DefaultShipLife = 10

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
