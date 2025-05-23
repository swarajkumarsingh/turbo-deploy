SCRIPT_FOLDER_NAME = commands

run:
	docker-compose up --build

runa:
	docker-compose -f docker-compose-alpine.yml up --build

build:
	docker build -t turbo-deploy . && docker run -p 8080:8080 turbo-deploy

start:
	docker run -p 8080:8080 turbo-deploy

compose:
	docker compose build && docker compose up

down:
	docker compose down

dev:
	nodemon --exec go run main.go

install:
	go mod tidy

run_prometheus:
	docker run -d -p 9090:9090 -v ./prometheus.yml:/etc/prometheus prom/prometheus

run_grafana:
	docker run -d --name=grafana -p 3000:3000 grafana/grafana-enterprise

gen:
	cd $(SCRIPT_FOLDER_NAME) && \
	go run *.go $n
	cd ..

deploy: 
	echo "TODO"

test: 
	echo "TODO"

.PHONY: build run logs dockerstop
.SILENT: build run logs dockerstop