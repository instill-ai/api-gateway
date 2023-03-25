# api-gateway

This repository maintains the [KrakenD](https://www.krakend.io) API gateway configuration file `krakend.json`.

## KrakenD

KrakenD is a binary executable processing the configuration file `krakend.json` for the API Gateway. 

The current used KrakenD version is `2.3.1` with Go `1.19.3` and Alpine `3.16`

## Local dev

On the local machine, clone `vdp` repository in your workspace, move to the repository folder, and launch all dependent microservices:
```bash
$ git clone https://github.com/instill-ai/vdp.git
$ cd vdp
$ make dev PROFILE=api-gateway
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
$ cd plugin && go build -buildmode=plugin -o grpc-proxy.so /api-gateway/plugin/server/grpc && cd .. # compile the KrakenD grpc-proxy plugin
$ make config # generate KrakenD configuration file
$ krakend run -c krakend.json
```

### CI/CD

- **push** to the `main` branch will trigger
  - the **`Create Release Candidate PR`** workflow, which will create and keep a PR to the `rc` branch up-to-date with respect to the `main` branch using [create-pull-request](github.com/peter-evans/create-pull-request) (commit message contains `release` string will be skipped), and
  - the **`Release Please`** workflow, which will create and update a PR with respect to the up-to-date `main` branch using [release-please-action](https://github.com/google-github-actions/release-please-action).
- **pull_request** to the `rc` branch will trigger the **`Integration Test`** workflow, which will run the integration test using the `:latest` images of **all** components.
- **push** to the `rc` branch will trigger
  - the **`Integration Test`** workflow, which will build the `:rc` image and run the integration test using the `:rc` image of all components, and
- Once the release PR is merged to the `main` branch, the [release-please-action](https://github.com/google-github-actions/release-please-action) will tag and release a version correspondingly.

The latest images are published to Docker Hub [repository](https://hub.docker.com/r/instill/api-gateway) at each CI/CD step.

## License

See the [LICENSE](./LICENSE) file for licensing information.
