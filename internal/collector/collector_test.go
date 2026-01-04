package collector

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/victorarias/immich-prometheus-exporter/internal/immich"
)

func TestCollector_Collect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/jobs":
			json.NewEncoder(w).Encode(immich.JobsResponse{
				"thumbnailGeneration": {
					JobCounts:   immich.JobCounts{Active: 3, Waiting: 10, Failed: 2, Delayed: 1},
					QueueStatus: immich.QueueStatus{IsActive: true, IsPaused: false},
				},
			})
		case "/api/server/statistics":
			json.NewEncoder(w).Encode(immich.StatisticsResponse{
				Photos: 5000,
				Videos: 1000,
				Usage:  100000000000,
				UsageByUser: []immich.UserUsage{
					{UserName: "alice", Photos: 5000, Videos: 1000, Usage: 100000000000},
				},
			})
		case "/api/server/storage":
			json.NewEncoder(w).Encode(immich.StorageResponse{
				DiskSize:      1000000000000,
				DiskUse:       500000000000,
				DiskAvailable: 500000000000,
			})
		}
	}))
	defer server.Close()

	client := immich.NewClient(server.URL, "test-key")
	collector := New(client)

	// Verify metrics are emitted
	count := testutil.CollectAndCount(collector)
	if count == 0 {
		t.Error("expected metrics to be collected")
	}

	// Check specific metric values
	expected := `
		# HELP immich_library_photos Total photos
		# TYPE immich_library_photos gauge
		immich_library_photos 5000
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "immich_library_photos"); err != nil {
		t.Errorf("unexpected metric value: %v", err)
	}

	expected = `
		# HELP immich_storage_total_bytes Total disk size
		# TYPE immich_storage_total_bytes gauge
		immich_storage_total_bytes 1e+12
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "immich_storage_total_bytes"); err != nil {
		t.Errorf("unexpected metric value: %v", err)
	}
}

func TestCollector_Describe(t *testing.T) {
	client := immich.NewClient("http://localhost", "test-key")
	collector := New(client)

	ch := make(chan *prometheus.Desc, 100)
	collector.Describe(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	// Should have all metric descriptors
	if count < 15 {
		t.Errorf("expected at least 15 metric descriptors, got %d", count)
	}
}

func TestCollector_ScrapeSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/jobs":
			json.NewEncoder(w).Encode(immich.JobsResponse{})
		case "/api/server/statistics":
			json.NewEncoder(w).Encode(immich.StatisticsResponse{})
		case "/api/server/storage":
			json.NewEncoder(w).Encode(immich.StorageResponse{})
		}
	}))
	defer server.Close()

	client := immich.NewClient(server.URL, "test-key")
	collector := New(client)

	expected := `
		# HELP immich_scrape_success Whether scrape succeeded (1=yes, 0=no)
		# TYPE immich_scrape_success gauge
		immich_scrape_success 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "immich_scrape_success"); err != nil {
		t.Errorf("unexpected metric value: %v", err)
	}
}

func TestCollector_ScrapeFailure(t *testing.T) {
	// Server that always returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := immich.NewClient(server.URL, "test-key")
	collector := New(client)

	expected := `
		# HELP immich_scrape_success Whether scrape succeeded (1=yes, 0=no)
		# TYPE immich_scrape_success gauge
		immich_scrape_success 0
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "immich_scrape_success"); err != nil {
		t.Errorf("unexpected metric value: %v", err)
	}
}
