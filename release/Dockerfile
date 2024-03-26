ARG ARCH=amd64
FROM --platform=linux/${ARCH} node:alpine

RUN apk add dumb-init

COPY /app/package.json /app/package.json
COPY /app/server.js /app/server.js
COPY /app/public /app/public

RUN npm install --prefix /app

EXPOSE 5000

WORKDIR "/app"
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "./server.js"]
