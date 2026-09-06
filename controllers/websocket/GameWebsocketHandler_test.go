package websocket

import (
	"sync"
	"testing"
	"time"
)

// resetBroadcastState puts the package globals back to a known-empty state so
// each test starts clean (they are package level, shared between tests).
func resetBroadcastState(t *testing.T) {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()
	Players = make(map[string]*playerData)
	playersToSend = make(map[string]*playerData)
	userConnections = make(map[string]*safeConn)
	newBullets = []*bulletData{}
	bulletsToRemove = []string{}
	killsList = []*playerHitData{}
	npcs = make(map[string]npcData)
	lastBroadcastTime = 0
}

// With no players connected, ships-npc stays connected and keeps simulating,
// so bullets and kills keep arriving. Those used to accumulate forever
// because the zero-player early return skipped the drain, and were then
// delivered as one enormous frame to the first player to join.
func TestCollectBroadcastDrainsBuffersWithNoPlayers(t *testing.T) {
	resetBroadcastState(t)

	// Simulate a few ticks worth of NPC activity with nobody playing.
	for i := 0; i < 100; i++ {
		mu.Lock()
		newBullets = append(newBullets, &bulletData{})
		bulletsToRemove = append(bulletsToRemove, "bullet")
		killsList = append(killsList, &playerHitData{})
		mu.Unlock()

		collectBroadcast()
	}

	mu.Lock()
	defer mu.Unlock()
	if len(newBullets) > 1 || len(bulletsToRemove) > 1 || len(killsList) > 1 {
		t.Fatalf("buffers grew unbounded with no players: newBullets=%d bulletsToRemove=%d killsList=%d",
			len(newBullets), len(bulletsToRemove), len(killsList))
	}
}

// Even with nobody playing, the NPC controller must still get a heartbeat
// every 2s carrying an empty activePlayerIds: that is the only way it learns
// the last player left. Without it, ships-npc chases and shoots a ghost.
func TestCollectBroadcastHeartbeatsWithNoPlayers(t *testing.T) {
	resetBroadcastState(t)

	// A fresh tick right after the previous broadcast sends nothing...
	mu.Lock()
	lastBroadcastTime = time.Now().UnixMilli()
	mu.Unlock()
	if payload, _ := collectBroadcast(); payload != nil {
		t.Fatal("expected no payload immediately after a broadcast")
	}

	// ...but once the 2s window has elapsed, a heartbeat goes out.
	mu.Lock()
	lastBroadcastTime = time.Now().UnixMilli() - 2001
	mu.Unlock()
	payload, _ := collectBroadcast()
	if payload == nil {
		t.Fatal("expected a heartbeat 2s after the last broadcast with no players")
	}
	ids, ok := payload["activePlayerIds"].([]string)
	if !ok || len(ids) != 0 {
		t.Fatalf("expected an empty activePlayerIds in the heartbeat, got %#v", payload["activePlayerIds"])
	}
}

// The payload keeps the slices it was built with, so draining must replace
// them rather than truncate them - otherwise the next tick would mutate a
// payload that is still being marshalled and written to the sockets.
func TestCollectBroadcastPayloadNotMutatedByNextTick(t *testing.T) {
	resetBroadcastState(t)

	mu.Lock()
	Players["p1"] = &playerData{SocketId: "p1"}
	newBullets = append(newBullets, &bulletData{})
	mu.Unlock()

	payload, _ := collectBroadcast()
	if payload == nil {
		t.Fatal("expected a payload when there is a player and a new bullet")
	}
	sent := payload["newBullets"].([]*bulletData)
	if len(sent) != 1 {
		t.Fatalf("expected 1 bullet in the payload, got %d", len(sent))
	}

	mu.Lock()
	newBullets = append(newBullets, &bulletData{}, &bulletData{})
	mu.Unlock()

	if len(sent) != 1 {
		t.Fatalf("payload slice was mutated by a later tick: now %d bullets", len(sent))
	}
}

// buildBackgroundCards is called from every client's reader goroutine on
// join. It used to write the global map behind a `len() > 0` check, which is
// a fatal "concurrent map writes" under -race and in production.
func TestBuildBackgroundCardsIsConcurrencySafe(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buildBackgroundCards()
		}()
	}
	wg.Wait()

	if len(BackgroundCards) != 5 {
		t.Fatalf("expected 5 background card columns, got %d", len(BackgroundCards))
	}
	for x, col := range BackgroundCards {
		if len(col) != 5 {
			t.Fatalf("column %d is half-built: %d rows", x, len(col))
		}
	}
}
