# Developing

You'll only need [mise](https://mise.jdx.dev).

```shell
make clean
make generate
make test
make build
```


## Adding a new demo

```shell
make demo/add
```

This will prompt you for info to add a new kustomize overlay for this demo.
In practice it adds a new folder in `kustomize/overlays` with the necessary files to deploy a new demo.

Also when you run `make k8s` (part of `make generate`) it will to `kustomize build` for each demo.
This can be useful to have demos as a single yaml.

#### Some advice for building a good demo

- Make it self-contained
- Keep it simple (don't demo too many things at once)
- Don't hesitate to have many files
