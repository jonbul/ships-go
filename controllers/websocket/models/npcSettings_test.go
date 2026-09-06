package models

import (
	"encoding/json"
	"strings"
	"testing"
)

// The toggle must survive the full round trip the admin panel drives:
// browser JSON -> NpcSettingsData -> Sanitized() -> npcConfig on the wire.
// A `false` in particular has to stay on the wire, or turning the toggle
// back off would leave ships-npc on its previous value.
func TestFightEachOtherSurvivesTheRoundTrip(t *testing.T) {
	for _, want := range []bool{true, false} {
		var in NpcSettingsData
		body := `{"enemyShips":3,"enemyShipLife":10,"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,"blackHoleSpawnPeriodSec":30,"enemyShipsFightEachOther":`
		if want {
			body += "true}"
		} else {
			body += "false}"
		}
		if err := json.Unmarshal([]byte(body), &in); err != nil {
			t.Fatal(err)
		}
		if in.EnemyShipsFightEachOther != want {
			t.Fatalf("decode: got %v want %v", in.EnemyShipsFightEachOther, want)
		}

		out := in.Sanitized()
		if out.EnemyShipsFightEachOther != want {
			t.Fatalf("Sanitized() dropped the toggle: got %v want %v", out.EnemyShipsFightEachOther, want)
		}

		encoded, err := json.Marshal(NpcConfigData{EventName: "npcConfig", Settings: out})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(encoded), `"enemyShipsFightEachOther"`) {
			t.Fatalf("field omitted from the wire (omitempty?): %s", encoded)
		}
	}
}
