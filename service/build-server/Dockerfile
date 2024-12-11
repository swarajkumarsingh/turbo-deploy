FROM ubuntu:focal

RUN apt-get update && \
    apt-get install -y curl git && \
    apt-get upgrade -y

RUN curl -sL https://deb.nodesource.com/setup_22.x | bash - && \
    apt-get install -y nodejs


WORKDIR /home/app

COPY main.sh main.sh
COPY script.js script.js
COPY vulnerableCommands.js vulnerableCommands.js
COPY package*.json ./

RUN npm install

RUN chmod +x main.sh

ENTRYPOINT [ "/home/app/main.sh" ]
