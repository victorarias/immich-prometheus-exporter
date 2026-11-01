FROM alpine:3.24

RUN apk add --no-cache ca-certificates

COPY immich-prometheus-exporter /immich-prometheus-exporter

EXPOSE 8080

ENTRYPOINT ["/immich-prometheus-exporter"]
