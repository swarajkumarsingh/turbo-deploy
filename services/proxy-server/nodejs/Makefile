run:
	docker build -t reverse-proxy .
	docker rm -f node-app
	docker run -d --name node-app  -p 3333:3333 node-app

eb_create:
	eb create turbo-deploy-proxy-server-env --timeout 20

eb_deploy:
	eb deploy

local:
	pm2 start index.js