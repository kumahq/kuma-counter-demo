const express = require('express')
const Redis = require("ioredis");

const KEY = "counter"
const app = express();
const PORT = 5000;

var version = process.env.APP_VERSION || "1.0";
var color = process.env.APP_COLOR || "#efefef";

function getClient() {
  var client = new Redis({
    port: parseInt(process.env.REDIS_PORT) || 6379,
    host: process.env.REDIS_HOST || "127.0.0.1",
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
  client.incr(KEY, function (err, result) {
    client.quit();
    if (err) {
      console.log(err);
      res.send({err:true});
    } else {
      res.send({counter: result, err: err});
    }
  });
});

app.get('/counter', function(req, res){
  var client = getClient();
  client.get(KEY, function(err, result) {
    client.quit();
    if (err) {
      console.log(err);
      res.send({err:true});
    } else {
      res.send({counter: result, err: err});
    }
  });
});

app.get('/version', function(req, res) {
  res.send({
    version: version,
    color: color
  });
});

app.listen(PORT, function(){
  console.log("Server running on port %s", PORT);
});