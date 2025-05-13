package internal

import (
	"fmt"
	"go-a2s-reporter/internal/helper" // Assuming this exists
	"log"
	"log/slog" // Added for consistency
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var initOnce sync.Once // Keep this for prometheus init

var (
	reg = prometheus.NewRegistry()
	// ... (keep your metric definitions) ...
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
			Help: "The current map on the server (labelled). Value is 1 if this server/map is active.",
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

	// To store the last known good info for clearing old labels
	lastKnownGoodServerName string
	lastKnownGoodMapName    string
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
	for queryResult := range DueForUpdate { // Now receives QueryResult
		info := queryResult.Info
		err := queryResult.Err

		if err != nil {
			slog.Error("Server query failed", "err", err)
			serverStatus.Set(0)
			// Clear potentially stale simple gauges
			serverPlayerCount.Set(0)
			serverMaxPlayerCount.Set(0)
			serverBots.Set(0)
			serverPasswordSet.Set(0)
			serverVac.Set(0)

			// For GaugeVecs, if we had a previously known good state, clear those specific labels
			if lastKnownGoodServerName != "" {
				serverPlayerCountWithServerName.DeleteLabelValues(lastKnownGoodServerName)
				if lastKnownGoodMapName != "" { // Map name might not always be present
					serverMap.DeleteLabelValues(lastKnownGoodServerName, lastKnownGoodMapName)
				}
			}
			lastKnownGoodServerName = "" // No current good state
			lastKnownGoodMapName = ""
			continue
		}

		if info == nil {
			// This case should ideally be covered by err != nil,
			// but good to handle defensively.
			slog.Warn("Received nil server info without an error, treating as server down.")
			serverStatus.Set(0)
			serverPlayerCount.Set(0)
			serverMaxPlayerCount.Set(0)
			serverBots.Set(0)
			serverPasswordSet.Set(0)
			serverVac.Set(0)
			if lastKnownGoodServerName != "" {
				serverPlayerCountWithServerName.DeleteLabelValues(lastKnownGoodServerName)
				if lastKnownGoodMapName != "" {
					serverMap.DeleteLabelValues(lastKnownGoodServerName, lastKnownGoodMapName)
				}
			}
			lastKnownGoodServerName = ""
			lastKnownGoodMapName = ""
			continue
		}

		// Server is UP and we have info
		serverStatus.Set(1)

		// If server name or map changed, delete old GaugeVec entries
		if lastKnownGoodServerName != "" && lastKnownGoodServerName != info.Name {
			serverPlayerCountWithServerName.DeleteLabelValues(lastKnownGoodServerName)
		}
		if lastKnownGoodServerName != "" && lastKnownGoodMapName != "" &&
			(lastKnownGoodServerName != info.Name || lastKnownGoodMapName != info.Map) {
			serverMap.DeleteLabelValues(lastKnownGoodServerName, lastKnownGoodMapName)
		}

		// Update metrics with new info
		serverPlayerCount.Set(float64(info.Players))
		serverMaxPlayerCount.Set(float64(info.MaxPlayers))
		serverPlayerCountWithServerName.WithLabelValues(info.Name).Set(float64(info.Players))
		serverBots.Set(float64(info.Bots))
		serverMap.WithLabelValues(info.Name, info.Map).Set(1)      // Use 1 to indicate current map
		serverPasswordSet.Set(helper.BoolToFloat(info.Visibility)) // Assuming Visibility means password-protected
		serverVac.Set(helper.BoolToFloat(info.VAC))

		// Store current info as last known good for the next cycle
		lastKnownGoodServerName = info.Name
		lastKnownGoodMapName = info.Map
	}
}

func Serve(port int) {
	slog.Info("Starting Prometheus metrics server", "port", port)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	if err != nil {
		slog.Error("HTTP server ListenAndServe error", "err", err)
		log.Fatal(err) // Original log.Fatal is fine here for a top-level failure
	}
}

// Assume GetEnvironmentVars is defined elsewhere and accessible
// type EnvironmentVars struct { /* ... */ }
// func GetEnvironmentVars() EnvironmentVars { /* ... */ }
