build:
	docker build -t build-server .

run:
	docker run --env-file .env build-server

start:
	make build && make run

test:
	node script.js

testt:
	rm -rf node_modules && rm -rf output

install:
	npm install

uninstall:
	rm -rf node_modules
