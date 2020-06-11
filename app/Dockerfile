FROM node:alpine

COPY package.json /app/package.json
COPY server.js /app/server.js
COPY public /app/public

RUN npm install --prefix /app

EXPOSE 5000

STOPSIGNAL SIGQUIT

CMD ["npm", "start", "--prefix", "/app"]