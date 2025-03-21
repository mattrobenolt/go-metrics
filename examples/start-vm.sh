#!/usr/bin/env bash

url="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v1.113.0/victoria-metrics-darwin-arm64-v1.113.0.tar.gz"
bin="./victoria-metrics-prod"
if [ ! -x "$bin" ]; then
    wget "$url" -O victoria-metrics.tar.gz
    tar xf victoria-metrics.tar.gz
    rm victoria-metrics.tar.gz
fi

exec "$bin" -promscrape.config prometheus.yaml
