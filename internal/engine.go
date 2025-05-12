package internal

import (
	"fmt"
	"github.com/rumblefrog/go-a2s"
	"log"
	"log/slog"
	"sync"
	"time"
)

var (
	client      *a2s.Client
	onceEngine  sync.Once
	tickerOnce  sync.Once
	latestQuery *a2s.ServerInfo
)

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
		ticker := time.NewTicker(time.Millisecond * time.Duration(GetEnvironmentVars().QueryInterval))
		go func() {
			defer ticker.Stop()
			for range ticker.C {
				info, err := forceFetchLatestServerInfo()
				if err != nil {
					slog.Error("Unable to fetch server info!", "err", err)
					latestQuery = nil
					continue
				}
				if info == nil {
					slog.Warn("Expected a2s.ServerInfo but received nil!")
				}
				latestQuery = info

				// Signal to update
				DueForUpdate <- struct{}{}
			}
		}()
	})
}

func forceFetchLatestServerInfo() (*a2s.ServerInfo, error) {
	return client.QueryInfo()
}

func FetchLatestServerInfo() *a2s.ServerInfo {
	return latestQuery
}
