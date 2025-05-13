package internal

import (
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/rumblefrog/go-a2s"
)

// QueryResult will hold the outcome of a server query attempt
type QueryResult struct {
	Info *a2s.ServerInfo
	Err  error
}

var (
	client     *a2s.Client
	onceEngine sync.Once
	tickerOnce sync.Once
	// latestQuery *a2s.ServerInfo // REMOVE this global for the reporter's direct use
)

// DueForUpdate now carries the QueryResult
var DueForUpdate chan QueryResult = make(chan QueryResult)

func init() {
	onceEngine.Do(func() {
		// Fetch environment variables
		env := GetEnvironmentVars()
		// Instantiate client for the server instance
		a2sClient, err := a2s.NewClient(
			fmt.Sprintf("%s:%d", env.GameAddress, env.GamePort),
			a2s.SetMaxPacketSize(uint32(env.QueryMaxPacketSize)),
			a2s.TimeoutOption(time.Millisecond*time.Duration(env.QueryTimeout)),
		)
		if err != nil {
			slog.Error("Encountered a fatal error creating A2S client", "err", err)
			log.Fatal(err)
		}
		client = a2sClient
		// Start query cycle
		initTicker()
	})
}

func initTicker() {
	tickerOnce.Do(func() {
		env := GetEnvironmentVars() // Get env vars once for the ticker setup
		ticker := time.NewTicker(time.Millisecond * time.Duration(env.QueryInterval))

		// Perform an initial query right away so metrics are available on startup
		go func() {
			slog.Info("Performing initial server query...")
			info, err := forceFetchLatestServerInfo()
			DueForUpdate <- QueryResult{Info: info, Err: err}
		}()

		go func() {
			defer ticker.Stop()
			for range ticker.C {
				info, err := forceFetchLatestServerInfo()
				// No longer update global latestQuery here for the reporter.
				// Send the result directly.
				DueForUpdate <- QueryResult{Info: info, Err: err}
			}
		}()
	})
}

func forceFetchLatestServerInfo() (*a2s.ServerInfo, error) {
	if client == nil { // Should not happen due to init, but good for robustness
		return nil, fmt.Errorf("A2S client is not initialized")
	}
	return client.QueryInfo()
}
