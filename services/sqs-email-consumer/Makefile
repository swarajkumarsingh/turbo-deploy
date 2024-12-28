.PHONY: build run logs dockerstop
.SILENT: build run logs dockerstop

run:
	docker compose up --build

build:
	echo "tody"

eb_create:
	eb create turbo-deploy-email-sqs-consumer-env --timeout 20

eb_deploy:
	eb deploy

install:
	go mod tidy

run_prometheus:
	docker run -d -p 9090:9090 -v ./prometheus.yml:/etc/prometheus prom/prometheus

run_grafana:
	docker run -d --name=grafana -p 3000:3000 grafana/grafana-enterprise
