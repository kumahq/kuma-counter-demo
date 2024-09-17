const express = require('express')
const Redis = require("ioredis");
const timers = require("timers/promises");
const util = require("util");
const path = require('path');

const COUNTER_KEY = "counter"
const ZONE_KEY = "zone"
const PORT = 5000;

const app = express();

app.set('view engine', 'ejs');
app.use((req, res, next) => {
  const prefix = req.headers['x-forwarded-prefix'] || ''; // Default to empty string if no prefix
  req.prefix = prefix; // Make prefix accessible in request
  next();
});


var version = process.env.APP_VERSION || "1.0";
var color = process.env.APP_COLOR || "#efefef";

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

// Set the static folder (for serving images, CSS, and client-side JS)
app.use(express.static(path.join(__dirname, 'public')));

// Route to render the index.ejs file
app.get('/', (req, res) => {
  res.render('index', { prefix: req.prefix }); // This will render views/index.ejs
});


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
