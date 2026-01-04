package immich

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	// Normalize trailing slash
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// JobsResponse maps queue name to job queue status
type JobsResponse map[string]JobQueue

type JobQueue struct {
	JobCounts   JobCounts   `json:"jobCounts"`
	QueueStatus QueueStatus `json:"queueStatus"`
}

type JobCounts struct {
	Active    int `json:"active"`
	Waiting   int `json:"waiting"`
	Failed    int `json:"failed"`
	Delayed   int `json:"delayed"`
	Paused    int `json:"paused"`
	Completed int `json:"completed"`
}

type QueueStatus struct {
	IsActive bool `json:"isActive"`
	IsPaused bool `json:"isPaused"`
}

type StatisticsResponse struct {
	Photos      int64       `json:"photos"`
	Videos      int64       `json:"videos"`
	Usage       int64       `json:"usage"`
	UsageByUser []UserUsage `json:"usageByUser"`
}

type UserUsage struct {
	UserName string `json:"userName"`
	Photos   int64  `json:"photos"`
	Videos   int64  `json:"videos"`
	Usage    int64  `json:"usage"`
}

type StorageResponse struct {
	DiskSize            int64   `json:"diskSizeRaw"`
	DiskUse             int64   `json:"diskUseRaw"`
	DiskAvailable       int64   `json:"diskAvailableRaw"`
	DiskUsagePercentage float64 `json:"diskUsagePercentage"`
}

func (c *Client) doRequest(path string, result interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

func (c *Client) GetJobs() (JobsResponse, error) {
	var result JobsResponse
	if err := c.doRequest("/api/jobs", &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) GetStatistics() (*StatisticsResponse, error) {
	var result StatisticsResponse
	if err := c.doRequest("/api/server/statistics", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetStorage() (*StorageResponse, error) {
	var result StorageResponse
	if err := c.doRequest("/api/server/storage", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Ping checks connectivity to Immich by calling the jobs endpoint
func (c *Client) Ping() error {
	_, err := c.GetJobs()
	return err
}
