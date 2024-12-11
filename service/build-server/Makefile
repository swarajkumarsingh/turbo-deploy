build:
	docker build -t build-server .

run:
	docker run --env-file .env build-server

test:
	node script.js

install:
	npm install

uninstall:
	rm -rf node_modules