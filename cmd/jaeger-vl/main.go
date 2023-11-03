package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/envflag"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/httpserver"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/logger"
	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"github.com/z-anshun/jaeger-vmlogs/app/vlselect"
	"github.com/z-anshun/jaeger-vmlogs/app/vlstorage"
	"github.com/z-anshun/jaeger-vmlogs/cmd/jaeger-vl/store"
	"gopkg.in/yaml.v3"
)

var (
	configPath = flag.String("config", "config.yaml", "A path to the plugin's configuration file")
)

func main() {

	flag.CommandLine.SetOutput(os.Stdout)
	envflag.Parse()

	vlstorage.Init()
	vlselect.Init()

	logger.Init()
	hcLog := hclog.New(&hclog.LoggerOptions{
		Name: "jaeger-vl",
		// If this is set to e.g. Warn, the debug logs are never sent to Jaeger even despite
		// --grpc-storage-plugin.log-level=debug
		Level:      hclog.Trace,
		JSONFormat: true,
	})

	// init config
	cfgFile, err := os.ReadFile(filepath.Clean(*configPath))
	if err != nil {
		hcLog.Error("Could not read config file", "config", configPath, "error", err)
		os.Exit(1)
	}
	var cfg store.Configuration
	err = yaml.Unmarshal(cfgFile, &cfg)
	if err != nil {
		hcLog.Error("Could not parse config file", "error", err)
	}
	cfg.SetDefaults()

	go httpserver.Serve(cfg.HttpListenAddr, false, requestHandler)

	str := store.NewStore(hcLog, &cfg)
	var pluginServices shared.PluginServices

	pluginServices.Store = str
	logger.Infof("starting VictoriaLogs")
	grpc.Serve(&pluginServices)

}

func requestHandler(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path == "/" {
		if r.Method != http.MethodGet {
			return false
		}
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<h2>Single-node VictoriaLogs</h2></br>")
		fmt.Fprintf(w, "See docs at <a href='https://docs.victoriametrics.com/VictoriaLogs/'>https://docs.victoriametrics.com/VictoriaLogs/</a></br>")
		fmt.Fprintf(w, "Useful endpoints:</br>")
		httpserver.WriteAPIHelp(w, [][2]string{
			{"select/vmui", "Web UI for VictoriaLogs"},
			{"metrics", "available service metrics"},
			{"flags", "command-line flags"},
		})
		return true
	}

	if vlselect.RequestHandler(w, r) {
		return true
	}
	return true
}
