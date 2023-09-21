# api-gateway

[![Integration Test](https://github.com/instill-ai/api-gateway/actions/workflows/integration-test.yml/badge.svg)](https://github.com/instill-ai/api-gateway/actions/workflows/integration-test.yml)

This repository maintains the [KrakenD](https://www.krakend.io) API gateway configuration file `krakend.json`.

## KrakenD

KrakenD is a binary executable processing the configuration file `krakend.json` for the API Gateway.

The current used KrakenD version is `2.3.1` with Go `1.19.3` and Alpine `3.16`

## Local dev

On the local machine, clone the desired project repository in your workspace either [base](https://github.com/instill-ai/base), [vdp](https://github.com/instill-ai/vdp) or [model](https://github.com/instill-ai/model), then move to the repository folder, and launch all dependent microservices:

```bash
$ git clone https://github.com/instill-ai/<project-name-here>.git
$ cd <project-name-here>
$ make latest PROFILE=api-gateway
```

Clone `api-gateway` repository in your workspace and move to the repository folder:

```bash
$ git clone https://github.com/instill-ai/api-gateway.git
$ cd api-gateway
```

### Build the dev image

```bash
$ make build
```

### Run the dev container

```bash
$ make dev
```

Now, you have the Go project set up in the container, in which you can compile and run the binaries together with the integration test in each container shell.

### Run the api-gateway server

```bash
# Enter api-gateway container
$ docker exec -it api-gateway /bin/bash

# In the api-gateway container
$ cd grpc_proxy_plugin && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugin/grpc-proxy.so /api-gateway/grpc_proxy_plugin/client && cd /api-gateway # compile the KrakenD grpc-proxy plugin
$ cd multi_auth_plugin && go build -buildmode=plugin -o /usr/local/lib/krakend/plugin/multi-auth.so /api-gateway/multi_auth_plugin/server && cd /api-gateway # compile the KrakenD multi-auth plugin
$ make config # generate KrakenD configuration file
$ krakend run -c krakend.json
```

### CI/CD

- **pull_request** to the `main` branch will trigger the **`Integration Test`** workflow running the integration test using the image built on the PR head branch.
- **push** to the `main` branch will trigger
  - the **`Integration Test`** workflow building and pushing the `:latest` image on the `main` branch, following by running the integration test, and
  - the **`Release Please`** workflow, which will create and update a PR with respect to the up-to-date `main` branch using [release-please-action](https://github.com/google-github-actions/release-please-action).

Once the release PR is merged to the `main` branch, the [release-please-action](https://github.com/google-github-actions/release-please-action) will tag and release a version correspondingly.

The images are pushed to Docker Hub [repository](https://hub.docker.com/r/instill/api-gateway).

## Contributing

Please refer to the [Contributing Guidelines](./.github/CONTRIBUTING.md) for more details.

## License

See the [LICENSE](./LICENSE) file for licensing information.
