# Immich Prometheus Exporter

A complete Prometheus exporter for [Immich](https://immich.app) exposing all available metrics: job queues, library statistics, and storage.

## Installation

### Docker

```bash
docker run -d \
  -e IMMICH_URL=http://immich-server:2283 \
  -e IMMICH_API_KEY=your-api-key \
  -p 8080:8080 \
  ghcr.io/victorarias/immich-prometheus-exporter:latest
```

### Docker Compose

```yaml
services:
  immich-exporter:
    image: ghcr.io/victorarias/immich-prometheus-exporter:latest
    environment:
      - IMMICH_URL=http://immich-server:2283
      - IMMICH_API_KEY=your-api-key
    ports:
      - "8080:8080"
```

### Binary

Download from [Releases](https://github.com/victorarias/immich-prometheus-exporter/releases), then:

```bash
export IMMICH_URL=http://localhost:2283
export IMMICH_API_KEY=your-api-key
./immich-prometheus-exporter
```

## Configuration

| Environment Variable | Required | Default | Description |
|---------------------|----------|---------|-------------|
| `IMMICH_URL` | Yes | - | Immich server URL (e.g., `http://localhost:2283`) |
| `IMMICH_API_KEY` | Yes | - | API key from Immich (Admin → API Keys) |
| `LISTEN_ADDRESS` | No | `:8080` | Address to listen on |

## Endpoints

| Path | Description |
|------|-------------|
| `/metrics` | Prometheus metrics |
| `/health` | Health check (verifies Immich connectivity) |

## Metrics

### Job Queue Metrics

Per-queue job counts with `queue` label:

```
immich_job_active{queue="thumbnailGeneration"} 3
immich_job_waiting{queue="ocr"} 1451
immich_job_failed{queue="thumbnailGeneration"} 2
immich_job_delayed{queue="faceDetection"} 0
immich_job_paused{queue="thumbnailGeneration"} 0
immich_job_completed{queue="thumbnailGeneration"} 0
immich_queue_active{queue="thumbnailGeneration"} 1
immich_queue_paused{queue="thumbnailGeneration"} 0
```

**Available queues:** `thumbnailGeneration`, `metadataExtraction`, `videoConversion`, `smartSearch`, `duplicateDetection`, `faceDetection`, `facialRecognition`, `sidecar`, `library`, `migration`, `backgroundTask`, `search`, `notifications`, `backupDatabase`, `ocr`, `workflow`, `storageTemplateMigration`

### Library Metrics

```
immich_library_photos 4881
immich_library_videos 1363
immich_library_bytes 100695517856

# Per-user breakdown
immich_user_photos{user="alice"} 3000
immich_user_videos{user="alice"} 800
immich_user_bytes{user="alice"} 60000000000
```

### Storage Metrics

```
immich_storage_total_bytes 7851124719616
immich_storage_used_bytes 142510915584
immich_storage_available_bytes 7708613804032
immich_storage_usage_percent 2.11
```

### Exporter Metrics

```
immich_scrape_duration_seconds 0.045
immich_scrape_success 1
immich_exporter_build_info{version="1.0.0",commit="abc123",date="2024-01-01"} 1
```

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'immich'
    static_configs:
      - targets: ['immich-exporter:8080']
```

## Example Alerts

```yaml
groups:
  - name: immich
    rules:
      - alert: ImmichJobQueueBacklog
        expr: immich_job_waiting > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Immich job queue backlog"
          description: "Queue {{ $labels.queue }} has {{ $value }} waiting jobs"

      - alert: ImmichFailedJobs
        expr: immich_job_failed > 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Immich has failed jobs"
          description: "Queue {{ $labels.queue }} has {{ $value }} failed jobs"

      - alert: ImmichExporterDown
        expr: immich_scrape_success == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Immich exporter cannot reach Immich"
```

## Requirements

- Immich API key with admin privileges (for full statistics)
- Network access to Immich server

## License

GPLv3
