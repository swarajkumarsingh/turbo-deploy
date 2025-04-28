run:
	docker build -t reverse-proxy .
	docker rm -f reverse-proxy
	docker run --name reverse-proxy  -p 3333:3333 reverse-proxy

eb_create:
	eb create turbo-deploy-proxy-server-env --timeout 20

eb_deploy:
	eb deploy

local:
	pm2 start index.js