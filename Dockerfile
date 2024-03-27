ARG ARCH=amd64
FROM --platform=linux/${ARCH} node:alpine

RUN apk add dumb-init

COPY /app/package.json /app/package.json
RUN npm install --prefix /app

COPY /app/public /app/public
COPY /app/server.js /app/server.js

EXPOSE 5000

WORKDIR "/app"
ENTRYPOINT ["dumb-init", "--"]
CMD ["node", "./server.js"]
