FROM node:18-alpine AS build

WORKDIR /app

COPY package*.json ./
RUN npm install --only=production

COPY . .

FROM node:18-alpine

RUN npm install -g pm2

WORKDIR /app

COPY --from=build /app /app

EXPOSE 3333

CMD ["pm2-runtime", "index.js"]