.PHONY: build run logs dockerstop
.SILENT: build run logs dockerstop

run:
	docker compose build
	docker compose up

build:
	echo "tody"

install:
	go mod tidy

run_prometheus:
	docker run -d -p 9090:9090 -v ./prometheus.yml:/etc/prometheus prom/prometheus

run_grafana:
	docker run -d --name=grafana -p 3000:3000 grafana/grafana-enterprise