const express = require('express')
const Redis = require("ioredis");
const timers = require("timers/promises");
const util = require("util");

const COUNTER_KEY = "counter"
const ZONE_KEY = "zone"
const PORT = 5000;

const app = express();

const version = process.env.APP_VERSION || "1.0";
const color = process.env.APP_COLOR || "#FAFAFA";

function getClient() {
  const host = process.env.REDIS_HOST || "127.0.0.1";
  const port = parseInt(process.env.REDIS_PORT) || 6379;
  const ip_version = parseInt(process.env.IP_VERSION) || 4;
  console.log("Connecting to Redis at %s:%d", host, port);
  const client = new Redis({
    port: port,
    host: host,
    family: ip_version,
    autoResendUnfulfilledCommands: false,
    autoResubscribe: false,
    enableOfflineQueue: true,
    maxRetriesPerRequest: null,
    reconnectOnError: function (err) {
      return false;
    },
    retryStrategy: function(times) {
      return false;
    }
  });
  client.on("error", function() {
    // Ignore
  });
  return client;
}

const sleep = util.promisify(setTimeout);
const delayMiddleware = function (req, res, next) {
  const delay = parseInt(req.header("x-set-response-delay-ms") || 0, 10)
  sleep(delay).then(() => {
    next()
  })
}

const statusCodeMiddleware = function (req, res, next) {
  const statusCode = parseInt(req.header("x-set-response-status-code") || 200, 10)
  res.status(statusCode)
  next()
}

app.use(delayMiddleware)
app.use(statusCodeMiddleware)
app.use('/', express.static('public'));

app.post('/increment', function(req, res){
  const client = getClient();
  client.incr(COUNTER_KEY, function (err, counter_result) {
    if (err) {
      console.log(err);
      res.send({err:true});
    } else {
      client.get(ZONE_KEY, function(err, zone_result) {
        client.quit();
        if (err) {
          console.log(err);
          res.send({err:true});
        } else {
          if (counter_result == null) {
            counter_result = 0;
          }
          res.send({counter: counter_result, zone: zone_result, err: err});
        }
      });
    }
  });
});

app.delete('/counter', function(req, res){
  const client = getClient();
  client.del(COUNTER_KEY, function(err) {
    if (err) {
      console.log(err);
      res.send({err:true});
    } else {
      client.get(ZONE_KEY, function(err, zone_result) {
        client.quit();
        if (err) {
          console.log(err);
          res.send({err:true});
        } else {
          res.send({counter: 0, zone: zone_result, err: err});
        }
      });
    }
  });
});

app.get('/counter', function(req, res){
  const client = getClient();
  client.get(COUNTER_KEY, function(err, counter_result) {
    if (err) {
      console.log(err);
      res.send({err:true});
    } else {
      client.get(ZONE_KEY, function(err, zone_result) {
        client.quit();
        if (err) {
          console.log(err);
          res.send({err:true});
        } else {
          if (counter_result == null) {
            counter_result = 0;
          }
          res.send({counter: counter_result, zone: zone_result, err: err});
        }
      });
    }
  });
});

app.get('/version', function(req, res) {
  res.send({
    version: version,
    color: color
  });
});

const server = app.listen(PORT, function(){
  console.log("Server running on port %s", PORT);
});

const shutdown = async (event) => {
  console.log('%s signal received: wait 1s to ensure this endpoint is dropped and shutting down', event)
  await timers.setTimeout(1000);
  await util.promisify(server.close.bind(server))();
};

process.on('SIGTERM', shutdown)
process.on('SIGINT', shutdown)
