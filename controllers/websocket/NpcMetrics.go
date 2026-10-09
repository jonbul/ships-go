package websocket

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ships/controllers/websocket/models"
)

// This file holds the last resource sample ships-npc sent, so the
// Prometheus collector in the controllers package can publish it. ships-npc
// has no HTTP server and is not reachable from outside, so it pushes its
// numbers up the websocket it already has and ships-go, which *is* scraped,
// re-exports them. See NpcMetricsData for the fields.

// npcMetricsTimeout is how long a sample stays publishable. ships-npc
// reports every few seconds, so anything this old means it has stopped
// talking - the process is gone, wedged, or its connection is broken.
//
// Without a timeout the collector would keep serving the last sample
// forever, and a dashboard would show a dead service sitting at a steady,
// entirely fictional CPU figure. Going stale is the honest answer.
const npcMetricsTimeout = 30 * time.Second

var (
	npcMetricsMu sync.RWMutex
	npcMetrics   models.NpcMetricsData
	npcMetricsAt time.Time
)

// storeNpcMetrics records a sample from the NPC controller.
func storeNpcMetrics(m models.NpcMetricsData) {
	npcMetricsMu.Lock()
	defer npcMetricsMu.Unlock()
	npcMetrics, npcMetricsAt = m, time.Now()
}

// clearNpcMetrics drops the last sample, called when the NPC controller
// disconnects. Waiting for the timeout would leave up to half a minute of
// metrics from a process that is already known to be gone.
func clearNpcMetrics() {
	npcMetricsMu.Lock()
	defer npcMetricsMu.Unlock()
	npcMetrics, npcMetricsAt = models.NpcMetricsData{}, time.Time{}
}

// snapshotNpcMetrics returns the most recent sample, and whether it is
// still fresh enough to publish.
func snapshotNpcMetrics() (models.NpcMetricsData, bool) {
	npcMetricsMu.RLock()
	defer npcMetricsMu.RUnlock()
	if npcMetricsAt.IsZero() || time.Since(npcMetricsAt) > npcMetricsTimeout {
		return models.NpcMetricsData{}, false
	}
	return npcMetrics, true
}

// npcMetricsCollector publishes the resource usage ships-npc reports over
// the websocket as part of ships-go's own /metrics output.
//
// It is a custom collector rather than a set of package-level Gauges that
// the websocket handler writes into, for two reasons:
//
//   - Staleness. Gauges keep their last value forever, so when ships-npc
//     dies its metrics would freeze at whatever they were and the dashboard
//     would show a healthy service that no longer exists. Collecting on
//     demand means a scrape reflects what is true at scrape time:
//     ships_npc_up drops to 0 and the rest simply stop being reported.
//   - Types. CPU time is a counter (it only grows, and is meant to be
//     rate()'d); everything else is a gauge. A collector can say so; a bag
//     of Gauges cannot.
type npcMetricsCollector struct {
	up          *prometheus.Desc
	cpuSeconds  *prometheus.Desc
	cpuPercent  *prometheus.Desc
	resident    *prometheus.Desc
	heap        *prometheus.Desc
	heapSys     *prometheus.Desc
	goroutines  *prometheus.Desc
	npcs        *prometheus.Desc
	tickSeconds *prometheus.Desc
}

// NewNpcMetricsCollector builds it. Exported so PrometheusController can
// register it alongside the collectors describing this process.
func NewNpcMetricsCollector() prometheus.Collector {
	return &npcMetricsCollector{
		up: prometheus.NewDesc("ships_npc_up",
			"1 if the ships-npc controller is connected and reporting, 0 otherwise.", nil, nil),
		cpuSeconds: prometheus.NewDesc("ships_npc_cpu_seconds_total",
			"Total user + system CPU time consumed by the ships-npc process.", nil, nil),
		cpuPercent: prometheus.NewDesc("ships_npc_cpu_percent",
			"CPU used by ships-npc since the previous report, as a percentage of one core.", nil, nil),
		resident: prometheus.NewDesc("ships_npc_memory_resident_bytes",
			"Resident set size of the ships-npc process.", nil, nil),
		heap: prometheus.NewDesc("ships_npc_memory_heap_bytes",
			"Bytes of live heap objects in the ships-npc process.", nil, nil),
		heapSys: prometheus.NewDesc("ships_npc_memory_heap_sys_bytes",
			"Heap memory the ships-npc process has reserved from the OS.", nil, nil),
		goroutines: prometheus.NewDesc("ships_npc_goroutines",
			"Goroutines running in the ships-npc process.", nil, nil),
		npcs: prometheus.NewDesc("ships_npc_simulated_npcs",
			"NPCs currently being simulated by ships-npc.", nil, nil),
		tickSeconds: prometheus.NewDesc("ships_npc_tick_seconds",
			"Average duration of one ships-npc simulation tick, including the update it sends.", nil, nil),
	}
}

func (c *npcMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.cpuSeconds
	ch <- c.cpuPercent
	ch <- c.resident
	ch <- c.heap
	ch <- c.heapSys
	ch <- c.goroutines
	ch <- c.npcs
	ch <- c.tickSeconds
}

func (c *npcMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	metrics, ok := snapshotNpcMetrics()

	// ships_npc_up is always reported, precisely so there is something to
	// alert on when everything else disappears.
	up := 0.0
	if ok {
		up = 1
	}
	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, up)
	if !ok {
		return
	}

	ch <- prometheus.MustNewConstMetric(c.cpuSeconds, prometheus.CounterValue, metrics.CpuSeconds)
	ch <- prometheus.MustNewConstMetric(c.cpuPercent, prometheus.GaugeValue, metrics.CpuPercent)
	// 0 means ships-npc could not measure it, not that it is using no
	// memory, so the series is omitted rather than reported as a flat zero
	// that would read like a real measurement.
	if metrics.ResidentBytes > 0 {
		ch <- prometheus.MustNewConstMetric(c.resident, prometheus.GaugeValue, metrics.ResidentBytes)
	}
	ch <- prometheus.MustNewConstMetric(c.heap, prometheus.GaugeValue, metrics.HeapBytes)
	ch <- prometheus.MustNewConstMetric(c.heapSys, prometheus.GaugeValue, metrics.HeapSysBytes)
	ch <- prometheus.MustNewConstMetric(c.goroutines, prometheus.GaugeValue, float64(metrics.Goroutines))
	ch <- prometheus.MustNewConstMetric(c.npcs, prometheus.GaugeValue, float64(metrics.Npcs))
	ch <- prometheus.MustNewConstMetric(c.tickSeconds, prometheus.GaugeValue, metrics.TickSeconds)
}
