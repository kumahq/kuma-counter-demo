# Kuma Counter Demo

[![Run in Insomnia}](https://insomnia.rest/images/run.svg)](https://insomnia.rest/run/?label=Kuma-count-demo&uri=https%3A%2F%2Fgithub.com%2Fkumahq%2Fkuma-counter-demo%2Fblob%2Fmaster%2Fopenapi.yaml)

[![][kuma-logo]][kuma-url]

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/kumahq/kuma/blob/master/LICENSE)
[![Slack](https://img.shields.io/badge/Slack-4A154B?logo=slack)](https://join.slack.com/t/kuma-mesh/shared_invite/zt-1rcll3y6t-DkV_CAItZUoy0IvCwQ~jlQ)
[![Twitter](https://img.shields.io/twitter/follow/KumaMesh.svg?style=social&label=Follow)](https://twitter.com/intent/follow?screen_name=KumaMesh)

Welcome to a sample application that demonstrates the [Kuma](https://github.com/kumahq/kuma) service mesh in action. Kuma is designed to work across Kubernetes and VMs environments, with support for multi-zone deployments across many different clusters, data centers, and clouds.

To learn more about Kuma, see [the Kuma website](https://kuma.io).

Kuma is a CNCF Sandbox project.

## Introduction

The application consists of the same app instantiated differently to simulate 2 services:

- A `demo-app` service that presents a web application that allows us to increment a numeric counter
- A `kv` service which simulates a database.


```mermaid
flowchart LR
browser

subgraph kuma-mesh
edge-gateway
demo-app(demo-app :5050)
kv(kv :5050)
end
edge-gateway --> demo-app
demo-app --> kv
browser --> edge-gateway
```

The `demo-app` service presents a browser interface that listens on port `5050`.

You can set the zone key on the kv `curl -v -XPOST -d '{"value":"zone-1"}' localhost:5050/api/key-value/zone -H 'content-type: application/json'` where `localhost:5050` is your kv service.

## Run the application

Follow the [getting-started](https://kuma.io/docs/latest/quickstart/kubernetes-demo/) on the Kuma docs.

### Environment Variables

We can configure the following environment variables when running `demo-app`:

* `KV_URL`: The address at which to contact the service. 
* `APP_VERSION`: Lets you change the version number displayed in the main page of `demo-app`. Default is `1.0`.
* `APP_COLOR`: Lets you change background color of the `demo-app` main page. Default is `#efefef`.

The `APP_VERSION` and `APP_COLOR` environment variables are handy when we want to create different versions of `demo-app` and get immediate visual feedback when routing across them.

[kuma-url]: https://kuma.io/
[kuma-logo]: https://kuma-public-assets.s3.amazonaws.com/kuma-logo-v2.png

## Modifying responses

### Adding delay to response

To add delay to response you need to set header `x-set-response-delay-ms`. Example:

```shell
curl localhost:5050/api/counter -XPOST -H "x-set-response-delay-ms: 5000"
```

### Enforcing response status code

To enforce response status code you need to set header `x-set-response-status-code`. Example:

```shell
curl localhost:5050/api/counter -XPOST -H "x-set-response-status-code: 503"
```

## Developing

You'll only need [mise](https://mise.jdx.dev).

```shell
make clean
make generate
make test
make build
```
