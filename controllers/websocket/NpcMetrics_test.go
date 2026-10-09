package websocket

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"ships/controllers/websocket/models"
)

func TestNpcMetricsRoundTripAndGoStale(t *testing.T) {
	t.Cleanup(clearNpcMetrics)
	clearNpcMetrics()

	if _, ok := snapshotNpcMetrics(); ok {
		t.Fatal("metrics were publishable before ships-npc reported anything")
	}

	storeNpcMetrics(models.NpcMetricsData{CpuPercent: 12.5, ResidentBytes: 4096, Npcs: 200})
	got, ok := snapshotNpcMetrics()
	if !ok {
		t.Fatal("a fresh sample was not publishable")
	}
	if got.CpuPercent != 12.5 || got.ResidentBytes != 4096 || got.Npcs != 200 {
		t.Fatalf("sample came back changed: %+v", got)
	}

	// A sample older than the timeout means ships-npc has stopped talking.
	// Publishing it anyway would show a dead service holding a steady, and
	// entirely fictional, CPU figure.
	npcMetricsMu.Lock()
	npcMetricsAt = time.Now().Add(-npcMetricsTimeout - time.Second)
	npcMetricsMu.Unlock()
	if _, ok := snapshotNpcMetrics(); ok {
		t.Fatal("a sample older than the timeout was still publishable")
	}
}

// The NPC controller going away has to take its metrics with it, rather
// than leaving up to a timeout's worth of readings from a process that is
// already known to be gone.
func TestNpcMetricsAreDroppedWhenTheControllerDisconnects(t *testing.T) {
	t.Cleanup(clearNpcMetrics)

	storeNpcMetrics(models.NpcMetricsData{CpuPercent: 40})
	if _, ok := snapshotNpcMetrics(); !ok {
		t.Fatal("sample was not stored")
	}

	clearNpcMetrics()
	if _, ok := snapshotNpcMetrics(); ok {
		t.Fatal("metrics survived the controller disconnecting")
	}
}

// gather renders the collector exactly as the /metrics endpoint would, so
// these assertions are made against the real exposition text - types and
// all - rather than against the collector's internals.
func gather(t *testing.T) string {
	t.Helper()

	reg := prometheus.NewRegistry()
	reg.MustRegister(NewNpcMetricsCollector())
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gathering: %v", err)
	}

	var out bytes.Buffer
	encoder := expfmt.NewEncoder(&out, expfmt.NewFormat(expfmt.TypeTextPlain))
	for _, family := range families {
		if err := encoder.Encode(family); err != nil {
			t.Fatalf("encoding %s: %v", family.GetName(), err)
		}
	}
	return out.String()
}

// With no ships-npc connected the scrape must still say so, and must not
// carry any of its numbers: a dashboard showing a dead service idling at a
// plausible CPU figure is worse than one showing nothing at all.
func TestNpcMetricsReportDownWhenNothingIsReporting(t *testing.T) {
	t.Cleanup(clearNpcMetrics)
	clearNpcMetrics()

	got := gather(t)
	if !strings.Contains(got, "ships_npc_up 0") {
		t.Fatalf("expected ships_npc_up 0, got:\n%s", got)
	}
	for _, name := range []string{"ships_npc_cpu_seconds_total", "ships_npc_memory_resident_bytes"} {
		if strings.Contains(got, name) {
			t.Fatalf("%s was published with no controller connected:\n%s", name, got)
		}
	}
}

func TestNpcMetricsArePublishedWithTheRightTypes(t *testing.T) {
	t.Cleanup(clearNpcMetrics)
	storeNpcMetrics(models.NpcMetricsData{
		CpuSeconds:    42.5,
		CpuPercent:    17.25,
		ResidentBytes: 1 << 20,
		HeapBytes:     1 << 19,
		HeapSysBytes:  1 << 21,
		Goroutines:    12,
		Npcs:          200,
		TickSeconds:   0.006,
	})

	got := gather(t)
	for _, want := range []string{
		"ships_npc_up 1",
		// CPU time only ever grows, so it has to be exported as a counter
		// for rate() to mean anything in Grafana.
		"# TYPE ships_npc_cpu_seconds_total counter",
		"ships_npc_cpu_seconds_total 42.5",
		"# TYPE ships_npc_cpu_percent gauge",
		"ships_npc_cpu_percent 17.25",
		"ships_npc_memory_resident_bytes 1.048576e+06",
		"ships_npc_goroutines 12",
		"ships_npc_simulated_npcs 200",
		"ships_npc_tick_seconds 0.006",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

// A resident size of 0 means ships-npc could not measure it (no /proc), not
// that it is using no memory, so it must not be published as a real zero.
func TestUnmeasurableResidentMemoryIsOmitted(t *testing.T) {
	t.Cleanup(clearNpcMetrics)
	storeNpcMetrics(models.NpcMetricsData{CpuSeconds: 1, ResidentBytes: 0})

	got := gather(t)
	if strings.Contains(got, "ships_npc_memory_resident_bytes") {
		t.Fatalf("an unmeasurable resident size was published:\n%s", got)
	}
	if !strings.Contains(got, "ships_npc_up 1") {
		t.Fatalf("the rest of the sample was dropped too:\n%s", got)
	}
}
