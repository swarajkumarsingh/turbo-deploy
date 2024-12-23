run:
	docker build -t reverse-proxy .
	docker rm -f node-app
	docker run -d --name node-app  -p 3333:3333 node-app

local:
	pm2 start index.js