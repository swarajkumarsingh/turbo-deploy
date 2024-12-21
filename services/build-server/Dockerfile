FROM node:alpine AS builder

WORKDIR /app

COPY package*.json ./

RUN npm install --production

FROM node:alpine

COPY --from=builder /app/node_modules ./  # Only copy dependencies
COPY main.sh main.sh
COPY *.js ./

RUN chmod +x main.sh

ENTRYPOINT [ "/app/main.sh" ]