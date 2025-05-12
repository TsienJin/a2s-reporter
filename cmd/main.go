package main

import (
	"fmt"
	"go-a2s-reporter/internal"
	"log/slog"
)

func main() {
	env := internal.GetEnvironmentVars()
	slog.Info(fmt.Sprintf("Starting A2S reporter on port %d...", env.ReporterPort))
	internal.Serve(env.ReporterPort)
}
