package internal

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go-a2s-reporter/internal/helper"
	"log"
	"net/http"
	"sync"
)

var (
	DueForUpdate chan struct{} = make(chan struct{})
	initOnce     sync.Once
)

var (
	reg = prometheus.NewRegistry()

	serverStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_status",
		Help: "1 if server is up, 0 if not.",
	})
	serverPlayerCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_player_count",
		Help: "Current number of players on the server",
	})
	serverMaxPlayerCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_max_player_count",
		Help: "Maximum number of players allowed by the server",
	})
	serverPlayerCountWithServerName = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "a2s_server_player_count_with_server_name",
			Help: "Map of server name and player count",
		},
		[]string{"server"},
	)
	serverBots = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_bots",
		Help: "Number of bots in the server.",
	})
	serverMap = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "a2s_server_map_info",
			Help: "The current map on the server (labelled).",
		},
		[]string{"server_name", "map"},
	)
	serverPasswordSet = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_password_set",
		Help: "1 if password protected, 0 if public.",
	})
	serverVac = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "a2s_server_vac_enabled",
		Help: "1 if VAC enabled, 0 if not.",
	})
)

func init() {
	initOnce.Do(func() {
		// Register all metrics
		reg.MustRegister(serverStatus)
		reg.MustRegister(serverPlayerCount)
		reg.MustRegister(serverMaxPlayerCount)
		reg.MustRegister(serverPlayerCountWithServerName)
		reg.MustRegister(serverBots)
		reg.MustRegister(serverMap)
		reg.MustRegister(serverPasswordSet)
		reg.MustRegister(serverVac)

		// Start monitoring for updates
		go listenForUpdate()
	})
}

func listenForUpdate() {
	for range DueForUpdate {
		info := FetchLatestServerInfo()

		if info == nil {
			serverStatus.Set(0)
			serverPlayerCount.Set(0)
			serverMaxPlayerCount.Set(0)
			serverPlayerCountWithServerName.Reset()
			serverBots.Set(0)
			serverMap.Reset()
			serverPasswordSet.Set(0)
			serverVac.Set(0)
			continue
		}

		serverStatus.Set(1)
		serverPlayerCount.Set(float64(info.Players))
		serverMaxPlayerCount.Set(float64(info.MaxPlayers))
		serverPlayerCountWithServerName.WithLabelValues(info.Name).Set(float64(info.Players))
		serverBots.Set(float64(info.Bots))
		serverMap.WithLabelValues(info.Name, info.Map).Set(1)
		serverPasswordSet.Set(helper.BoolToFloat(info.Visibility))
		serverVac.Set(helper.BoolToFloat(info.VAC))
	}
}

func Serve(port int) {
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
