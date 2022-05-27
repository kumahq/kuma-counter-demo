ARG ARCH=amd64
FROM --platform=linux/${ARCH} node:alpine

COPY /app/package.json /app/package.json
COPY /app/server.js /app/server.js
COPY /app/public /app/public

RUN npm install --prefix /app

EXPOSE 5000

STOPSIGNAL SIGQUIT

CMD ["npm", "start", "--prefix", "/app"]