FROM node:lts

RUN apt-get update \
  && apt-get install --no-install-recommends --yes \
    bind9-host \
    curl \
    dnsutils \
    iputils-ping \
    jq \
    net-tools \
    redis-tools \
    telnet \
    vim \
    openssh-client \
    traceroute \
  && rm -rf /var/lib/apt/lists/*

COPY package.json /app/package.json
COPY server.js /app/server.js
COPY public /app/public

RUN npm install --prefix /app

EXPOSE 5000

STOPSIGNAL SIGQUIT

CMD ["npm", "start", "--prefix", "/app"]