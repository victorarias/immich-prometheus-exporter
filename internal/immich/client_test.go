package immich

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient_NormalizesTrailingSlash(t *testing.T) {
	client := NewClient("http://localhost:2283/", "test-key")
	if client.baseURL != "http://localhost:2283" {
		t.Errorf("expected baseURL without trailing slash, got %s", client.baseURL)
	}
}

func TestGetJobs(t *testing.T) {
	expected := JobsResponse{
		"thumbnailGeneration": {
			JobCounts:   JobCounts{Active: 3, Waiting: 10, Failed: 2, Delayed: 0},
			QueueStatus: QueueStatus{IsActive: true, IsPaused: false},
		},
		"faceDetection": {
			JobCounts:   JobCounts{Active: 1, Waiting: 100, Failed: 0, Delayed: 5},
			QueueStatus: QueueStatus{IsActive: true, IsPaused: false},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/jobs" {
			t.Errorf("expected path /api/jobs, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header, got %s", r.Header.Get("x-api-key"))
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	result, err := client.GetJobs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["thumbnailGeneration"].JobCounts.Active != 3 {
		t.Errorf("expected 3 active jobs, got %d", result["thumbnailGeneration"].JobCounts.Active)
	}
	if result["faceDetection"].JobCounts.Waiting != 100 {
		t.Errorf("expected 100 waiting jobs, got %d", result["faceDetection"].JobCounts.Waiting)
	}
}

func TestGetStatistics(t *testing.T) {
	expected := StatisticsResponse{
		Photos: 5000,
		Videos: 1000,
		Usage:  100000000000,
		UsageByUser: []UserUsage{
			{UserName: "alice", Photos: 3000, Videos: 600, Usage: 60000000000},
			{UserName: "bob", Photos: 2000, Videos: 400, Usage: 40000000000},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/server/statistics" {
			t.Errorf("expected path /api/server/statistics, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	result, err := client.GetStatistics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Photos != 5000 {
		t.Errorf("expected 5000 photos, got %d", result.Photos)
	}
	if len(result.UsageByUser) != 2 {
		t.Errorf("expected 2 users, got %d", len(result.UsageByUser))
	}
}

func TestGetStorage(t *testing.T) {
	expected := StorageResponse{
		DiskSize:      1000000000000,
		DiskUse:       500000000000,
		DiskAvailable: 500000000000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/server/storage" {
			t.Errorf("expected path /api/server/storage, got %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	result, err := client.GetStorage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DiskSize != 1000000000000 {
		t.Errorf("expected 1TB disk size, got %d", result.DiskSize)
	}
}

func TestGetJobs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-key")
	result, err := client.GetJobs()
	if err == nil {
		t.Error("expected error for 401 response")
	}
	if result != nil {
		t.Error("expected nil result on error")
	}
}

func TestPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(JobsResponse{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key")
	if err := client.Ping(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPing_Error(t *testing.T) {
	client := NewClient("http://localhost:99999", "test-key")
	if err := client.Ping(); err == nil {
		t.Error("expected error for unreachable server")
	}
}
