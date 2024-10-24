# Kuma Counter Demo

[![][kuma-logo]][kuma-url]

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kumahq/kuma/blob/master/LICENSE)
[![Slack](https://img.shields.io/badge/Slack-4A154B?logo=slack)](https://join.slack.com/t/kuma-mesh/shared_invite/zt-1rcll3y6t-DkV_CAItZUoy0IvCwQ~jlQ)
[![Twitter](https://img.shields.io/twitter/follow/KumaMesh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=KumaMesh)

Welcome to a sample application that demonstrates the [Kuma](https://kuma.io) service mesh in action. Kuma is designed to work across Kubernetes and VMs environments, with support for multi-zone deployments across many different clusters, data centers, and clouds.

To learn more about Kuma, see [the Kuma docs](https://kuma.io/docs/latest).

Kuma is a CNCF Sandbox project.

## Introduction

The application consists of two services:

- A `demo-app` service that presents a web application that allows us to increment a numeric counter
- A `redis` service that stores the counter

<img width="861" alt="kuma-counter-demo" src="https://user-images.githubusercontent.com/964813/124640078-c5efce00-de41-11eb-9513-4e11b88ca64c.png">

The `demo-app` service presents a browser interface that listens on port `5000`. When it starts, it expects to find a `zone` key in Redis that specifies the name of the datacenter (or cluster) that the current `redis` instance belongs to. This name is then displayed in the `demo-app` GUI.

The `zone` key is purely static and arbitrary, but by having different `zone` values across different `redis` instances, we know at any given time from which Redis instance we are fetching/incrementing our counter when we route across a distributed environment across many zones, clusters and clouds.

### Run the application

To run the `kuma-counter-demo` follow one of these guides:

- [Kubernetes quickstart demo](https://kuma.io/docs/latest/quickstart/kubernetes-demo/)
- [Universal quickstart demo](https://kuma.io/docs/latest/quickstart/universal-demo/)

### Environment Variables

We can configure the following environment variables when running `demo-app`:

* `REDIS_HOST`: Determines the hostname to use when connecting to Redis. Default is `127.0.0.1`.
* `REDIS_PORT`: Determines the port to use when connecting to Redis. Default is `6379`.
* `APP_VERSION`: Lets you change the version number displayed in the main page of `demo-app`. Default is `1.0`.
* `APP_COLOR`: Lets you change background color of the `demo-app` main page. Default is `#efefef`.

The `APP_VERSION` and `APP_COLOR` environment variables are handy when we want to create different versions of `demo-app` and get immediate visual feedback when routing across them.

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png

## Modifying responses

### Adding delay to response

To add delay to response you need to set header `x-set-response-delay-ms`. Example:

```shell
curl localhost:5001/increment -XPOST -H "x-set-response-delay-ms: 5000"
```

### Enforcing response status code

To enforce response status code you need to set header `x-set-response-status-code`. Example:

```shell
curl localhost:5001/increment -XPOST -H "x-set-response-status-code: 503"
```