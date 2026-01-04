package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/victorarias/immich-prometheus-exporter/internal/collector"
	"github.com/victorarias/immich-prometheus-exporter/internal/immich"
)

// Build info set by ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	immichURL := os.Getenv("IMMICH_URL")
	apiKey := os.Getenv("IMMICH_API_KEY")
	listenAddr := os.Getenv("LISTEN_ADDRESS")

	if immichURL == "" || apiKey == "" {
		log.Fatal("IMMICH_URL and IMMICH_API_KEY environment variables are required")
	}

	if listenAddr == "" {
		listenAddr = ":8080"
	}

	log.Printf("Immich Prometheus Exporter %s (commit: %s, built: %s)", version, commit, date)

	client := immich.NewClient(immichURL, apiKey)
	coll := collector.New(client)

	// Register build info metric
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "immich",
			Subsystem: "exporter",
			Name:      "build_info",
			Help:      "Build information",
		},
		[]string{"version", "commit", "date"},
	)
	buildInfo.WithLabelValues(version, commit, date).Set(1)
	prometheus.MustRegister(buildInfo)
	prometheus.MustRegister(coll)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := client.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("unhealthy: " + err.Error()))
			return
		}
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	// Graceful shutdown
	done := make(chan bool, 1)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
		done <- true
	}()

	log.Printf("Listening on %s", listenAddr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}

	<-done
	log.Println("Stopped")
}
