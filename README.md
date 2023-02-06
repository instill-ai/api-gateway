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

The latest images will be published to Docker Hub [repository](https://hub.docker.com/r/instill/api-gateway) at release.

## License

See the [LICENSE](./LICENSE) file for licensing information.
