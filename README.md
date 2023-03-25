# api-gateway

[![rc integration-test](https://img.shields.io/github/actions/workflow/status/instill-ai/api-gateway/integration-test.yml?branch=rc&label=rc%20integration-test%20&logoColor=fff&logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTYiIGhlaWdodD0iMTYiIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KPHBhdGggZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiIGQ9Ik0wIDEuNzVDMCAwLjc4NCAwLjc4NCAwIDEuNzUgMEg1LjI1QzYuMjE2IDAgNyAwLjc4NCA3IDEuNzVWNS4yNUM3IDUuNzE0MTMgNi44MTU2MyA2LjE1OTI1IDYuNDg3NDQgNi40ODc0NEM2LjE1OTI1IDYuODE1NjMgNS43MTQxMyA3IDUuMjUgN0g0VjExQzQgMTEuMjY1MiA0LjEwNTM2IDExLjUxOTYgNC4yOTI4OSAxMS43MDcxQzQuNDgwNDMgMTEuODk0NiA0LjczNDc4IDEyIDUgMTJIOVYxMC43NUM5IDkuNzg0IDkuNzg0IDkgMTAuNzUgOUgxNC4yNUMxNS4yMTYgOSAxNiA5Ljc4NCAxNiAxMC43NVYxNC4yNUMxNiAxNC43MTQxIDE1LjgxNTYgMTUuMTU5MiAxNS40ODc0IDE1LjQ4NzRDMTUuMTU5MiAxNS44MTU2IDE0LjcxNDEgMTYgMTQuMjUgMTZIMTAuNzVDMTAuMjg1OSAxNiA5Ljg0MDc1IDE1LjgxNTYgOS41MTI1NiAxNS40ODc0QzkuMTg0MzcgMTUuMTU5MiA5IDE0LjcxNDEgOSAxNC4yNVYxMy41SDVDNC4zMzY5NiAxMy41IDMuNzAxMDcgMTMuMjM2NiAzLjIzMjIzIDEyLjc2NzhDMi43NjMzOSAxMi4yOTg5IDIuNSAxMS42NjMgMi41IDExVjdIMS43NUMxLjI4NTg3IDcgMC44NDA3NTIgNi44MTU2MyAwLjUxMjU2MyA2LjQ4NzQ0QzAuMTg0Mzc0IDYuMTU5MjUgMCA1LjcxNDEzIDAgNS4yNUwwIDEuNzVaTTEuNzUgMS41QzEuNjgzNyAxLjUgMS42MjAxMSAxLjUyNjM0IDEuNTczMjIgMS41NzMyMkMxLjUyNjM0IDEuNjIwMTEgMS41IDEuNjgzNyAxLjUgMS43NVY1LjI1QzEuNSA1LjM4OCAxLjYxMiA1LjUgMS43NSA1LjVINS4yNUM1LjMxNjMgNS41IDUuMzc5ODkgNS40NzM2NiA1LjQyNjc4IDUuNDI2NzhDNS40NzM2NiA1LjM3OTg5IDUuNSA1LjMxNjMgNS41IDUuMjVWMS43NUM1LjUgMS42ODM3IDUuNDczNjYgMS42MjAxMSA1LjQyNjc4IDEuNTczMjJDNS4zNzk4OSAxLjUyNjM0IDUuMzE2MyAxLjUgNS4yNSAxLjVIMS43NVpNMTAuNzUgMTAuNUMxMC42ODM3IDEwLjUgMTAuNjIwMSAxMC41MjYzIDEwLjU3MzIgMTAuNTczMkMxMC41MjYzIDEwLjYyMDEgMTAuNSAxMC42ODM3IDEwLjUgMTAuNzVWMTQuMjVDMTAuNSAxNC4zODggMTAuNjEyIDE0LjUgMTAuNzUgMTQuNUgxNC4yNUMxNC4zMTYzIDE0LjUgMTQuMzc5OSAxNC40NzM3IDE0LjQyNjggMTQuNDI2OEMxNC40NzM3IDE0LjM3OTkgMTQuNSAxNC4zMTYzIDE0LjUgMTQuMjVWMTAuNzVDMTQuNSAxMC42ODM3IDE0LjQ3MzcgMTAuNjIwMSAxNC40MjY4IDEwLjU3MzJDMTQuMzc5OSAxMC41MjYzIDE0LjMxNjMgMTAuNSAxNC4yNSAxMC41SDEwLjc1WiIgZmlsbD0id2hpdGUiLz4KPC9zdmc+Cg==)](https://github.com/instill-ai/api-gateway/actions/workflows/integration-test.yml?branch=rc&event=push)

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
- **push** to the `rc` branch will trigger the **`Integration Test`** workflow, which will build the `:rc` image and run the integration test using the `:rc` image of all components.
- Once the release PR is merged to the `main` branch, the [release-please-action](https://github.com/google-github-actions/release-please-action) will tag and release a version correspondingly.

The latest images are published to Docker Hub [repository](https://hub.docker.com/r/instill/api-gateway) at each CI/CD step.

## License

See the [LICENSE](./LICENSE) file for licensing information.
