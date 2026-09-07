package controllers

import (
	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ships/controllers/websocket"
)

var activePlayers = websocket.ActivePlayers

func RegisterPrometheusRoutes(router *gin.Engine) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		activePlayers,
		// The Go and process collectors above describe *this* process. The
		// NPC simulation runs in ships-npc, which has no HTTP server and is
		// not reachable to be scraped, so it pushes its own CPU and memory
		// up the websocket and they are re-exported here.
		websocket.NewNpcMetricsCollector(),
	)

	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	srv, _ := statsviz.NewServer(statsviz.Root("/vi/status"))
	ws := srv.Ws()
	index := srv.Index()

	router.GET("/vi/status/*filepath", func(c *gin.Context) {
		if c.Param("filepath") == "/ws" {
			ws(c.Writer, c.Request)
			return
		}
		index(c.Writer, c.Request)
	})
}
