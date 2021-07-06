# Kuma Counter Demo

[![][kuma-logo]][kuma-url]

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kumahq/kuma/blob/master/LICENSE)
[![Slack](https://chat.kuma.io/badge.svg)](https://chat.kuma.io/)
[![Twitter](https://img.shields.io/twitter/follow/KumaMesh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=KumaMesh)

Welcome to a sample application that demonstrates the [Kuma](https://kuma.io) service mesh in action. Kuma is a service mesh designed to work across both Kubernetes and VMs environments, with support for multi-zone deployments across many different clusters, data centers and clouds.

To learn more about Kuma, please check out Kuma's [repository](https://github.com/kumahq/kuma).

Kuma is a CNCF Sandbox project.

## Introduction

This demo application consists of two services: a `demo-app` service that presents a web application that allows us to increment a numeric counter stored in another `redis` service.

The `demo-app` service by default presents a browser interface that listens on port `5000` that we can use to increment our counters. When the `demo-app` starts, it expects to find a `zone` key in Redis that it is being used to determine the arbitrary name of the datacenter (or cluster) that the current `redis` instance belongs to, which is then displayed in the `demo-app` GUI.

The `zone` key is purely static and arbitrary, but by having different `zone` values across different `redis` instances, we know at any given time from which Redis instance we are fetching/incrementing our counter when we route across a distributed environment across many zones, clusters and clouds.

### Running the application

First, we need to run `redis` on the default port `6379` - and set a default `zone` name - with:

```sh
$ redis-server
$ redis-cli set zone local
```

Then, we can start our `demo-app` on the default port `5000` with:

```sh
$ npm start --prefix=app/
```

Finally, we can navigate to [`127.0.0.1:5000`](http://127.0.0.1:5000) and increment our counter!

### Running in Kubernetes

We provide two different YAML files for Kubernetes, one that installs the basic resources and another one that installs a v2 version of the frontend service with different colors (useful to demonstrate routing across multiple versions, for example).

The following command installs the basic demo resources in a `kuma-demo` namespace:

```sh
$ kubectl apply -f demo.yaml
```

Which we can then access on port `5000` with a simple port forwarding:

```sh
$ kubectl port-forward svc/demo-app -n kuma-demo 5000:5000
```

Finally, we can navigate to [`127.0.0.1:5000`](http://127.0.0.1:5000) and increment our counter!

#### Installing v2

To install the `v2` version of the demo application, we can apply the following YAML file instead, which will change the colors and version number of our frontend application:

```sh
$ kubectl apply -f demo-v2.yaml
```

By inspecting the file we can see that the `demo-v2.yaml` file simply sets the following environment variables - which are also available on VMs - to different values so that we have immediate visual feedback that a new version is runnning.

### Environment Variables

There are some environment variables that we can configure when running `demo-app`:

* `REDIS_HOST`: Determines the hostname to use when connecting to Redis. By default is `127.0.0.1`.
* `REDIS_PORT`: Determines the port to use when connecting to Redis. By default is `6379`.
* `APP_VERSION`: Allows to change the version number displayed in the main page of `demo-app`. By default is `1.0`.
* `APP_COLOR`: Allows to change background color of the `demo-app` main page. By default is `#efefef`.

The `APP_VERSION` and `APP_COLOR` environment variables are handy when we want to create different versions of `demo-app` and get immediate visual feedback when routing across them.
