package websocket

import (
	"log"
	"sync"

	"ships/controllers/websocket/models"
)

// NPC settings live here rather than in the admin controller because they
// have two readers with different lifecycles: the admin panel (HTTP) and
// the NPC controller connection (websocket). Keeping them next to the
// connection registry means a change can be pushed to ships-npc the moment
// it's saved, and a freshly (re)connected ships-npc can be told the
// current values immediately - so the two processes converge no matter
// which one restarted last, with no polling and no extra endpoint.
//
// Defaults are duplicated in ships-npc's defaultNpcSettings(); both sides
// must agree so behaviour is identical before an admin ever opens the panel.
var (
	npcSettingsMu sync.RWMutex
	npcSettings   = models.NpcSettingsData{
		EnemyShips:              1,
		EnemyShipLife:           10,
		EnemyShipSpeed:          20,
		EnemyShipFireRateMs:     500,
		MaxBlackHoles:           2,
		BlackHoleSpawnPeriodSec: 30,
		// Off by default: NPCs fighting each other changes the feel of the
		// game a lot, so it is opt-in from the admin panel.
		EnemyShipsFightEachOther: false,
	}
)

// GetNpcSettings returns the NPC settings currently in force, for the
// admin panel to render.
func GetNpcSettings() models.NpcSettingsData {
	npcSettingsMu.RLock()
	defer npcSettingsMu.RUnlock()
	return npcSettings
}

// SetNpcSettings stores new admin-chosen NPC settings and pushes them to
// ships-npc straight away, so the change takes effect on its next tick
// without restarting anything. Values are clamped so what's stored (and
// echoed back to the admin panel) is exactly what ships-npc will run with.
func SetNpcSettings(settings models.NpcSettingsData) {
	npcSettingsMu.Lock()
	npcSettings = settings.Sanitized()
	npcSettingsMu.Unlock()

	// Collect first, write after releasing mu: a slow or dead socket must
	// never block the game loop, which takes the same lock every tick.
	mu.Lock()
	npcConns := make([]*safeConn, 0, 1)
	for _, c := range userConnections {
		if c.isNpc.Load() {
			npcConns = append(npcConns, c)
		}
	}
	mu.Unlock()

	for _, c := range npcConns {
		sendNpcSettings(c)
	}
}

// sendNpcSettings pushes the current settings to one NPC controller
// connection.
func sendNpcSettings(conn *safeConn) {
	if err := conn.writeJSON(models.NpcConfigData{
		EventName: "npcConfig",
		Settings:  GetNpcSettings(),
	}); err != nil {
		log.Println("Failed to send npcConfig to " + conn.remoteAddr + ": " + err.Error())
	}
}
