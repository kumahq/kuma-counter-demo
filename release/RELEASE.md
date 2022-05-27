## How to release `kuma-demo` image?

Kuma supports `amd64` and `arm64` architecture. To build and push multi-platform image you there is script that support this. Run:

```bash
DOCKER_USERNAME=<your-user-name> DOCKER_API_KEY=<your-api-key> DRY_RUN=<true|false> ./release/docker.sh
```

If you want to push your images than you need to set environment `DRY_RUN=false`, by default is true.