package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

// Every cell of the attack matrix must survive the full round trip the
// admin panel drives: browser JSON -> NpcSettingsData -> Sanitized() ->
// npcConfig on the wire. A `false` in particular has to stay on the wire,
// or clearing a box would leave ships-npc on its previous value.
func TestAttackMatrixSurvivesTheRoundTrip(t *testing.T) {
	cells := []struct {
		field string
		get   func(NpcSettingsData) bool
	}{
		{"npcAttacksPlayers", func(s NpcSettingsData) bool { return s.NpcAttacksPlayers }},
		{"npcAttacksNpc", func(s NpcSettingsData) bool { return s.NpcAttacksNpc }},
		{"npcAttacksAi", func(s NpcSettingsData) bool { return s.NpcAttacksAi }},
		{"aiAttacksPlayers", func(s NpcSettingsData) bool { return s.AiAttacksPlayers }},
		{"aiAttacksNpc", func(s NpcSettingsData) bool { return s.AiAttacksNpc }},
		{"aiAttacksAi", func(s NpcSettingsData) bool { return s.AiAttacksAi }},
	}

	for _, want := range []bool{true, false} {
		body := `{"enemyShipController":"both","enemyShips":3,"aiShips":4,"shipLife":10,` +
			`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,"blackHoleSpawnPeriodSec":30`
		for _, cell := range cells {
			if want {
				body += `,"` + cell.field + `":true`
			} else {
				body += `,"` + cell.field + `":false`
			}
		}
		body += "}"

		var in NpcSettingsData
		if err := json.Unmarshal([]byte(body), &in); err != nil {
			t.Fatal(err)
		}
		out := in.Sanitized()

		encoded, err := json.Marshal(NpcConfigData{EventName: "npcConfig", Settings: out})
		if err != nil {
			t.Fatal(err)
		}
		for _, cell := range cells {
			if cell.get(in) != want {
				t.Fatalf("decode of %s: got %v want %v", cell.field, cell.get(in), want)
			}
			if cell.get(out) != want {
				t.Fatalf("Sanitized() dropped %s: got %v want %v", cell.field, cell.get(out), want)
			}
			if !strings.Contains(string(encoded), `"`+cell.field+`"`) {
				t.Fatalf("%s omitted from the wire (omitempty?): %s", cell.field, encoded)
			}
		}
	}
}

// Both fleets are capped at MaxNpcFleetSize each, independently. An
// unrecognised controller must not silently mean "none": that would empty
// the map because of a typo upstream.
func TestSanitizedClampsFleetsAndController(t *testing.T) {
	out := NpcSettingsData{
		EnemyShipController: "nonsense",
		EnemyShips:          9999,
		AiShips:             -3,
	}.Sanitized()

	if out.EnemyShipController != NpcControllerRule {
		t.Fatalf("unknown controller became %q, want the default", out.EnemyShipController)
	}
	if out.EnemyShips != MaxNpcFleetSize {
		t.Fatalf("rule fleet not clamped to %d: got %d", MaxNpcFleetSize, out.EnemyShips)
	}
	if out.AiShips != 0 {
		t.Fatalf("negative AI fleet not clamped to 0: got %d", out.AiShips)
	}

	for _, controller := range []string{NpcControllerNone, NpcControllerRule, NpcControllerAi, NpcControllerBoth} {
		got := NpcSettingsData{EnemyShipController: controller}.Sanitized().EnemyShipController
		if got != controller {
			t.Fatalf("valid controller %q was rewritten to %q", controller, got)
		}
	}
}

// contactDamage is the one setting that also has to reach players' browsers,
// so it travels twice: to ships-npc inside the NPC settings, and to every
// player as a gameSettings event. Both are false-by-default booleans, which
// is exactly the shape that goes missing silently, so both are pinned here.
func TestContactDamageSurvivesTheRoundTrip(t *testing.T) {
	for _, want := range []bool{true, false} {
		body := `{"enemyShipController":"rule","enemyShips":1,"aiShips":1,"shipLife":10,` +
			`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,` +
			`"blackHoleSpawnPeriodSec":30,"contactDamage":` + strconv.FormatBool(want) + `}`

		var in NpcSettingsData
		if err := json.Unmarshal([]byte(body), &in); err != nil {
			t.Fatal(err)
		}
		out := in.Sanitized()
		if out.ContactDamage != want {
			t.Fatalf("contactDamage %v did not survive Sanitized: got %v", want, out.ContactDamage)
		}

		wire, err := json.Marshal(out)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(wire), `"contactDamage":`+strconv.FormatBool(want)) {
			t.Fatalf("contactDamage %v missing from the wire: %s", want, wire)
		}

		settings, err := json.Marshal(GameSettingsData{EventName: "gameSettings", ContactDamage: want})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(settings), `"contactDamage":`+strconv.FormatBool(want)) {
			t.Fatalf("contactDamage %v missing from the player broadcast: %s", want, settings)
		}
	}
}

// Score-based ship scaling is opt-in, and like contactDamage it is enforced
// by the browser, so it has to reach players in the gameSettings event as
// well as surviving the admin panel's round trip.
func TestKillScalingIsOffByDefaultAndSurvivesTheRoundTrip(t *testing.T) {
	for _, want := range []bool{true, false} {
		body := `{"enemyShipController":"rule","enemyShips":1,"aiShips":1,"shipLife":10,` +
			`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,` +
			`"blackHoleSpawnPeriodSec":30,"killScaling":` + strconv.FormatBool(want) + `}`

		var in NpcSettingsData
		if err := json.Unmarshal([]byte(body), &in); err != nil {
			t.Fatal(err)
		}
		out := in.Sanitized()
		if out.KillScaling != want {
			t.Fatalf("killScaling %v did not survive Sanitized: got %v", want, out.KillScaling)
		}

		wire, err := json.Marshal(out)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(wire), `"killScaling":`+strconv.FormatBool(want)) {
			t.Fatalf("killScaling %v missing from the wire: %s", want, wire)
		}

		settings, err := json.Marshal(GameSettingsData{EventName: "gameSettings", KillScaling: want})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(settings), `"killScaling":`+strconv.FormatBool(want)) {
			t.Fatalf("killScaling %v missing from the player broadcast: %s", want, settings)
		}
	}

	// A settings blob from a panel that has never heard of this option must
	// leave it off rather than inheriting whatever the field defaults to.
	var absent NpcSettingsData
	if err := json.Unmarshal([]byte(`{"enemyShipController":"rule"}`), &absent); err != nil {
		t.Fatal(err)
	}
	if absent.Sanitized().KillScaling {
		t.Fatal("killScaling should default to off")
	}
}

// ShipSize is the one number here that ships-vue and ships-npc both act on,
// so it has to survive the round trip and reach both. Zero means "not set"
// - an older admin page, or a settings document written before the field
// existed - and has to become the default rather than the minimum, which
// would silently shrink every ship in the game.
func TestShipSizeRoundTripAndDefaults(t *testing.T) {
	baseBody := `{"enemyShipController":"rule","enemyShips":1,"aiShips":1,"shipLife":10,` +
		`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,` +
		`"blackHoleSpawnPeriodSec":30`

	cases := []struct {
		name string
		body string
		want int
	}{
		{"absent", baseBody + `}`, DefaultShipSize},
		{"zero", baseBody + `,"shipSize":0}`, DefaultShipSize},
		{"negative", baseBody + `,"shipSize":-10}`, DefaultShipSize},
		{"below the minimum", baseBody + `,"shipSize":1}`, MinShipSize},
		{"above the maximum", baseBody + `,"shipSize":100000}`, MaxShipSize},
		{"kept", baseBody + `,"shipSize":250}`, 250},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var in NpcSettingsData
			if err := json.Unmarshal([]byte(c.body), &in); err != nil {
				t.Fatal(err)
			}
			out := in.Sanitized()
			if out.ShipSize != c.want {
				t.Fatalf("shipSize: got %d want %d", out.ShipSize, c.want)
			}

			wire, err := json.Marshal(out)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(wire), `"shipSize":`+strconv.Itoa(c.want)) {
				t.Fatalf("shipSize missing from the ships-npc wire: %s", wire)
			}

			settings, err := json.Marshal(GameSettingsData{EventName: "gameSettings", ShipSize: out.ShipSize})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(settings), `"shipSize":`+strconv.Itoa(c.want)) {
				t.Fatalf("shipSize missing from the player broadcast: %s", settings)
			}
		})
	}
}

// The black hole duration follows the same rule as every other number here:
// clamped, and an unset value takes the default rather than the minimum,
// which would turn every black hole into a five-second blink.
func TestBlackHoleDurationRoundTripAndDefaults(t *testing.T) {
	baseBody := `{"enemyShipController":"rule","enemyShips":1,"aiShips":1,"shipLife":10,` +
		`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,` +
		`"blackHoleSpawnPeriodSec":30`

	cases := []struct {
		name string
		body string
		want int
	}{
		{"absent", baseBody + `}`, DefaultBlackHoleDurationSec},
		{"zero", baseBody + `,"blackHoleDurationSec":0}`, DefaultBlackHoleDurationSec},
		{"too short", baseBody + `,"blackHoleDurationSec":1}`, 5},
		{"too long", baseBody + `,"blackHoleDurationSec":999999}`, 86400},
		{"kept", baseBody + `,"blackHoleDurationSec":45}`, 45},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var in NpcSettingsData
			if err := json.Unmarshal([]byte(c.body), &in); err != nil {
				t.Fatal(err)
			}
			out := in.Sanitized()
			if out.BlackHoleDurationSec != c.want {
				t.Fatalf("blackHoleDurationSec: got %d want %d", out.BlackHoleDurationSec, c.want)
			}
			wire, err := json.Marshal(out)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(wire), `"blackHoleDurationSec":`+strconv.Itoa(c.want)) {
				t.Fatalf("blackHoleDurationSec missing from the ships-npc wire: %s", wire)
			}
		})
	}
}

// Ship life used to be an NPC-only setting (enemyShipLife), so raising it
// left the player on the 10 ships-vue hardcoded while every NPC got the new
// value - exactly the kind of asymmetry the game rules are meant to
// prevent. It is now one number for everyone, which means it has to reach
// the browsers in gameSettings as well as ships-npc in npcConfig.
func TestShipLifeReachesPlayersAndNpcs(t *testing.T) {
	baseBody := `{"enemyShipController":"rule","enemyShips":1,"aiShips":1,` +
		`"enemyShipSpeed":20,"enemyShipFireRateMs":500,"maxBlackHoles":2,` +
		`"blackHoleSpawnPeriodSec":30`

	cases := []struct {
		name string
		body string
		want float32
	}{
		{"absent", baseBody + `}`, DefaultShipLife},
		{"zero", baseBody + `,"shipLife":0}`, DefaultShipLife},
		{"negative", baseBody + `,"shipLife":-5}`, DefaultShipLife},
		{"kept", baseBody + `,"shipLife":100}`, 100},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var in NpcSettingsData
			if err := json.Unmarshal([]byte(c.body), &in); err != nil {
				t.Fatal(err)
			}
			out := in.Sanitized()
			if out.ShipLife != c.want {
				t.Fatalf("shipLife: got %v want %v", out.ShipLife, c.want)
			}

			want := `"shipLife":` + strconv.FormatFloat(float64(c.want), 'g', -1, 32)
			wire, err := json.Marshal(out)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(wire), want) {
				t.Fatalf("shipLife missing from the ships-npc wire: %s", wire)
			}

			settings, err := json.Marshal(GameSettingsData{EventName: "gameSettings", ShipLife: out.ShipLife})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(settings), want) {
				t.Fatalf("shipLife missing from the player broadcast: %s", settings)
			}
		})
	}
}
