FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o immich-prometheus-exporter ./cmd/immich-prometheus-exporter

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/immich-prometheus-exporter /immich-prometheus-exporter

EXPOSE 8080

ENTRYPOINT ["/immich-prometheus-exporter"]
