const express = require('express')
const Redis = require("ioredis");
const timers = require("timers/promises");
const util = require("util");

const COUNTER_KEY = "counter"
const ZONE_KEY = "zone"
const PORT = 5000;

const app = express();

var version = process.env.APP_VERSION || "1.0";
var color = process.env.APP_COLOR || "#efefef";

const gracefulShutdownSeconds = process.env.GRACEFUL_SHUTDOWN_SECONDS || "1";

function getClient() {
  var host = process.env.REDIS_HOST || "127.0.0.1";
  var port = parseInt(process.env.REDIS_PORT) || 6379;
  console.log("Connecting to Redis at %s:%d", host, port);
  var client = new Redis({
    port: port,
    host: host,
    family: 4,
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

app.use('/', express.static('public'));

app.post('/increment', function(req, res){
  var client = getClient();
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
  var client = getClient();
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
  var client = getClient();
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

app.get('/stream-counter', async function(req, res){
  var client = getClient();
  res.flushHeaders();
  while (true) {
      await util.promisify(setTimeout)(1000);
      try {
        const counter = await client.get(COUNTER_KEY);
        const zone = await client.get(ZONE_KEY);
        res.write(JSON.stringify({counter, zone, err: false}));
      } catch (err) {
        res.write(JSON.stringify({err: true}));
      }
      res.write("\n");
  }
  client.quit();
  res.end();
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
  const now = new Date();
  console.log('%s: %s signal received: wait %ss to ensure this endpoint is dropped and shutting down', now.toISOString(), event, gracefulShutdownSeconds);
  await timers.setTimeout(parseInt(gracefulShutdownSeconds)*1000);
  await util.promisify(server.close.bind(server))();
};

process.on('SIGTERM', shutdown)
process.on('SIGINT', shutdown)
