package controllers

import (
	"github.com/arl/statsviz"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterPrometheusRoutes(router *gin.Engine) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	srv, _ := statsviz.NewServer()
	router.GET("/status/*filepath", func(c *gin.Context) {
		if c.Param("filepath") == "/ws" {
			srv.Ws()(c.Writer, c.Request)
			return
		}
		srv.Index()(c.Writer, c.Request)
	})
}
