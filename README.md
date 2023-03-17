# Kuma Counter Demo

[![][kuma-logo]][kuma-url]

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kumahq/kuma/blob/master/LICENSE)
[![Slack](https://img.shields.io/badge/Slack-4A154B?logo=slack)](https://join.slack.com/t/kuma-mesh/shared_invite/zt-1rcll3y6t-DkV_CAItZUoy0IvCwQ~jlQ)
[![Twitter](https://img.shields.io/twitter/follow/KumaMesh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=KumaMesh)

Welcome to a sample application that demonstrates the [Kuma](https://kuma.io) service mesh in action. Kuma is designed to work across Kubernetes and VMs environments, with support for multi-zone deployments across many different clusters, data centers, and clouds.

To learn more about Kuma, see [the Kuma repository](https://github.com/kumahq/kuma).

Kuma is a CNCF Sandbox project.

## Introduction

The application consists of two services:

- A `demo-app` service that presents a web application that allows us to increment a numeric counter
- A `redis` service that stores the counter

<img width="861" alt="kuma-counter-demo" src="https://user-images.githubusercontent.com/964813/124640078-c5efce00-de41-11eb-9513-4e11b88ca64c.png">

The `demo-app` service presents a browser interface that listens on port `5000`. When it starts, it expects to find a `zone` key in Redis that specifies the name of the datacenter (or cluster) that the current `redis` instance belongs to. This name is then displayed in the `demo-app` GUI.

The `zone` key is purely static and arbitrary, but by having different `zone` values across different `redis` instances, we know at any given time from which Redis instance we are fetching/incrementing our counter when we route across a distributed environment across many zones, clusters and clouds.

### Run the application

1.  Run `redis`

    - (Kubernetes setup) on the default port `6379`  and set a default `zone` name:

    ```sh
    $ redis-server
    $ redis-cli set zone local
    ```

    - (Universal setup) on port `26379` and set a default `zone` name:

     ```sh
    $ redis-server --port 26379
    $ redis-cli -p 26379 set zone local
    ```

2.  Install and start `demo-app` on the default port `5000`:

    ```sh
    $ npm install --prefix=app/
    $ npm start --prefix=app/
    ```

3.  (Only for Kubernetess) Navigate to [`127.0.0.1:5000`](http://127.0.0.1:5000) and increment the counter!


### Run in Universal

1. First we need to generate two tokens:
 - One for redis
 - Second for node-app
To do so we need to run:

    ```sh
    $ kumactl generate dataplane-token --name=redis > kuma-token-redis
    $ kumactl generate dataplane-token --name=app > kuma-token-app
    ```

2. Then we need to generate two DPPs:
- For redis:

```sh
kuma-dp run \
   --cp-address=https://localhost:5678/ \
   --dns-enabled=false \
   --dataplane-token-file=kuma-token-redis \
   --dataplane="type: Dataplane
mesh: default
name: redis
networking:
  address: 127.0.0.1 # Or any public address (needs to be reachable by every other dp in the zone) 
  inbound:
    - port: 16379
      servicePort: 26379
      tags:
        kuma.io/service: redis
        kuma.io/protocol: tcp"
```


- And for app:

```sh
kuma-dp run \
  --cp-address=https://localhost:5678/ \
  --dns-enabled=false \
  --dataplane-token-file=kuma-token-app \
  --dataplane="type: Dataplane
mesh: default
name: app
networking:
  address: 127.0.0.1 # Or any public address (needs to be reachable by every other dp in the zone) 
  outbound:
    - port: 6379
      tags:
        kuma.io/service: redis
  inbound:
    - port: 15000
      servicePort: 5000
      tags:
        kuma.io/service: app
        kuma.io/protocol: http
  admin:
    port: 9902"
```

3.  Navigate to [`127.0.0.1:5000`](http://127.0.0.1:5000) and increment the counter!

### Run in Kubernetes

Two different YAML files are provided for Kubernetes:

- one that installs the basic resources
- one that installs a version of the frontend service with different colors (useful to demonstrate routing across multiple versions, for example)

1.  Install the basic demo resources in a `kuma-demo` namespace:

    ```sh
    $ kubectl apply -f demo.yaml
    ```

1.  Port-forward the service to the namespace on port `5000`:

    ```sh
    $ kubectl port-forward svc/demo-app -n kuma-demo 5000:5000
    ```

1.  Navigate to [`127.0.0.1:5000`](http://127.0.0.1:5000) and increment the counter!

#### `MeshGateway`

An additional YAML file `gateway.yaml` is provided for setting up a
[Kuma `MeshGateway`](https://kuma.io/docs/1.7.x/explore/gateway/#builtin) to
serve the demo app.

1.  Install the `MeshGateway` resources:

    ```sh
    $ kubectl apply -f gateway.yaml
    ```

1. If your k8s environment supports `LoadBalancer` `Services`, you can access
   the gateway via its external IP:

   ```
   $ curl $(kubectl get svc -n kuma-demo demo-app-gateway -o=jsonpath='{.status.loadBalancer.ingress[0].ip}')
   ```

   Otherwise you can access the gateway inside the cluster via its cluster IP:

   ```
   $ kubectl run curl-gateway --rm -i --tty --image nicolaka/netshoot --restart=OnFailure -- curl $(kubectl get svc -n kuma-demo demo-app-gateway -o=jsonpath='{.spec.clusterIP}')
   ```

### Run v2 (Kubernetes)

To install the `v2` version of the demo application, we can apply the following YAML file instead, which will change the colors and version number of our frontend application:

```sh
$ kubectl apply -f demo-v2.yaml
```

By inspecting the file we can see that the `demo-v2.yaml` file sets the following environment variables - which are also available on VMs - to different values so that we have immediate visual feedback that a new version is runnning.

### Environment Variables

We can configure the following environment variables when running `demo-app`:

* `REDIS_HOST`: Determines the hostname to use when connecting to Redis. Default is `127.0.0.1`.
* `REDIS_PORT`: Determines the port to use when connecting to Redis. Default is `6379`.
* `APP_VERSION`: Lets you change the version number displayed in the main page of `demo-app`. Default is `1.0`.
* `APP_COLOR`: Lets you change background color of the `demo-app` main page. Default is `#efefef`.

The `APP_VERSION` and `APP_COLOR` environment variables are handy when we want to create different versions of `demo-app` and get immediate visual feedback when routing across them.

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png
