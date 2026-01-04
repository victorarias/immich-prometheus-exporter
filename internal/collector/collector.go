package collector

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/victorarias/immich-prometheus-exporter/internal/immich"
)

const namespace = "immich"

type ImmichCollector struct {
	client *immich.Client

	// Job metrics (per-queue)
	jobActive    *prometheus.Desc
	jobWaiting   *prometheus.Desc
	jobFailed    *prometheus.Desc
	jobDelayed   *prometheus.Desc
	jobPaused    *prometheus.Desc
	jobCompleted *prometheus.Desc
	queueActive  *prometheus.Desc
	queuePaused  *prometheus.Desc

	// Library metrics
	libraryPhotos *prometheus.Desc
	libraryVideos *prometheus.Desc
	libraryBytes  *prometheus.Desc
	userPhotos    *prometheus.Desc
	userVideos    *prometheus.Desc
	userBytes     *prometheus.Desc

	// Storage metrics
	storageTotal        *prometheus.Desc
	storageUsed         *prometheus.Desc
	storageAvailable    *prometheus.Desc
	storageUsagePercent *prometheus.Desc

	// Exporter metrics
	scrapeDuration *prometheus.Desc
	scrapeSuccess  *prometheus.Desc
}

func New(client *immich.Client) *ImmichCollector {
	return &ImmichCollector{
		client: client,

		// Job metrics
		jobActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "active"),
			"Number of active jobs",
			[]string{"queue"}, nil,
		),
		jobWaiting: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "waiting"),
			"Number of waiting jobs",
			[]string{"queue"}, nil,
		),
		jobFailed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "failed"),
			"Number of failed jobs",
			[]string{"queue"}, nil,
		),
		jobDelayed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "delayed"),
			"Number of delayed jobs",
			[]string{"queue"}, nil,
		),
		jobPaused: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "paused"),
			"Number of paused jobs",
			[]string{"queue"}, nil,
		),
		jobCompleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "job", "completed"),
			"Number of completed jobs",
			[]string{"queue"}, nil,
		),
		queueActive: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "queue", "active"),
			"Whether queue is active (1=yes, 0=no)",
			[]string{"queue"}, nil,
		),
		queuePaused: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "queue", "paused"),
			"Whether queue is paused (1=yes, 0=no)",
			[]string{"queue"}, nil,
		),

		// Library metrics
		libraryPhotos: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "library", "photos"),
			"Total photos",
			nil, nil,
		),
		libraryVideos: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "library", "videos"),
			"Total videos",
			nil, nil,
		),
		libraryBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "library", "bytes"),
			"Total storage usage in bytes",
			nil, nil,
		),
		userPhotos: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "user", "photos"),
			"Photos per user",
			[]string{"user"}, nil,
		),
		userVideos: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "user", "videos"),
			"Videos per user",
			[]string{"user"}, nil,
		),
		userBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "user", "bytes"),
			"Storage per user in bytes",
			[]string{"user"}, nil,
		),

		// Storage metrics
		storageTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "total_bytes"),
			"Total disk size",
			nil, nil,
		),
		storageUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "used_bytes"),
			"Disk used",
			nil, nil,
		),
		storageAvailable: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "available_bytes"),
			"Disk available",
			nil, nil,
		),
		storageUsagePercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "storage", "usage_percent"),
			"Disk usage percentage",
			nil, nil,
		),

		// Exporter metrics
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "duration_seconds"),
			"Time taken to scrape",
			nil, nil,
		),
		scrapeSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "success"),
			"Whether scrape succeeded (1=yes, 0=no)",
			nil, nil,
		),
	}
}

func (c *ImmichCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobActive
	ch <- c.jobWaiting
	ch <- c.jobFailed
	ch <- c.jobDelayed
	ch <- c.jobPaused
	ch <- c.jobCompleted
	ch <- c.queueActive
	ch <- c.queuePaused
	ch <- c.libraryPhotos
	ch <- c.libraryVideos
	ch <- c.libraryBytes
	ch <- c.userPhotos
	ch <- c.userVideos
	ch <- c.userBytes
	ch <- c.storageTotal
	ch <- c.storageUsed
	ch <- c.storageAvailable
	ch <- c.storageUsagePercent
	ch <- c.scrapeDuration
	ch <- c.scrapeSuccess
}

func (c *ImmichCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	success := 1.0

	var wg sync.WaitGroup
	var jobsResp immich.JobsResponse
	var statsResp *immich.StatisticsResponse
	var storageResp *immich.StorageResponse
	var jobsErr, statsErr, storageErr error

	// Fetch all APIs in parallel
	wg.Add(3)

	go func() {
		defer wg.Done()
		jobsResp, jobsErr = c.client.GetJobs()
	}()

	go func() {
		defer wg.Done()
		statsResp, statsErr = c.client.GetStatistics()
	}()

	go func() {
		defer wg.Done()
		storageResp, storageErr = c.client.GetStorage()
	}()

	wg.Wait()

	// Process job metrics
	if jobsErr != nil {
		log.Printf("Error fetching jobs: %v", jobsErr)
		success = 0
	} else {
		for queueName, queue := range jobsResp {
			ch <- prometheus.MustNewConstMetric(c.jobActive, prometheus.GaugeValue, float64(queue.JobCounts.Active), queueName)
			ch <- prometheus.MustNewConstMetric(c.jobWaiting, prometheus.GaugeValue, float64(queue.JobCounts.Waiting), queueName)
			ch <- prometheus.MustNewConstMetric(c.jobFailed, prometheus.GaugeValue, float64(queue.JobCounts.Failed), queueName)
			ch <- prometheus.MustNewConstMetric(c.jobDelayed, prometheus.GaugeValue, float64(queue.JobCounts.Delayed), queueName)
			ch <- prometheus.MustNewConstMetric(c.jobPaused, prometheus.GaugeValue, float64(queue.JobCounts.Paused), queueName)
			ch <- prometheus.MustNewConstMetric(c.jobCompleted, prometheus.GaugeValue, float64(queue.JobCounts.Completed), queueName)
			ch <- prometheus.MustNewConstMetric(c.queueActive, prometheus.GaugeValue, boolToFloat(queue.QueueStatus.IsActive), queueName)
			ch <- prometheus.MustNewConstMetric(c.queuePaused, prometheus.GaugeValue, boolToFloat(queue.QueueStatus.IsPaused), queueName)
		}
	}

	// Process statistics metrics
	if statsErr != nil {
		log.Printf("Error fetching statistics: %v", statsErr)
		success = 0
	} else {
		ch <- prometheus.MustNewConstMetric(c.libraryPhotos, prometheus.GaugeValue, float64(statsResp.Photos))
		ch <- prometheus.MustNewConstMetric(c.libraryVideos, prometheus.GaugeValue, float64(statsResp.Videos))
		ch <- prometheus.MustNewConstMetric(c.libraryBytes, prometheus.GaugeValue, float64(statsResp.Usage))

		for _, user := range statsResp.UsageByUser {
			ch <- prometheus.MustNewConstMetric(c.userPhotos, prometheus.GaugeValue, float64(user.Photos), user.UserName)
			ch <- prometheus.MustNewConstMetric(c.userVideos, prometheus.GaugeValue, float64(user.Videos), user.UserName)
			ch <- prometheus.MustNewConstMetric(c.userBytes, prometheus.GaugeValue, float64(user.Usage), user.UserName)
		}
	}

	// Process storage metrics
	if storageErr != nil {
		log.Printf("Error fetching storage: %v", storageErr)
		success = 0
	} else {
		ch <- prometheus.MustNewConstMetric(c.storageTotal, prometheus.GaugeValue, float64(storageResp.DiskSize))
		ch <- prometheus.MustNewConstMetric(c.storageUsed, prometheus.GaugeValue, float64(storageResp.DiskUse))
		ch <- prometheus.MustNewConstMetric(c.storageAvailable, prometheus.GaugeValue, float64(storageResp.DiskAvailable))
		ch <- prometheus.MustNewConstMetric(c.storageUsagePercent, prometheus.GaugeValue, storageResp.DiskUsagePercentage)
	}

	// Exporter metrics
	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, duration)
	ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, success)
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
