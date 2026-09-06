package models

import (
	"encoding/json"
	"testing"
)

// The client sends "hide" and reads "hide" back; the field was tagged
// "hidden", so it never survived the round trip and a player waiting to
// respawn stayed drawn on everyone else's screen.
func TestPlayerDataRoundTripsTheFieldsTheClientSends(t *testing.T) {
	const in = `{"socketId":"s","x":-250,"y":949,"scale":0.18,"width":1000,"height":1000,"hide":true,"isDead":true,"shipId":"custom"}`

	var p PlayerData
	if err := json.Unmarshal([]byte(in), &p); err != nil {
		t.Fatal(err)
	}
	if !p.Hide || !p.IsDead || p.Width != 1000 || p.Height != 1000 || p.Scale != 0.18 {
		t.Fatalf("decoded %+v", p)
	}

	out, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(out, &wire); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]any{
		"hide": true, "width": 1000.0, "height": 1000.0, "scale": 0.18,
	} {
		if got, ok := wire[k]; !ok || got != want {
			t.Fatalf("wire[%q] = %v (present=%v), want %v", k, got, ok, want)
		}
	}
}
