# api-gateway

This repository maintains the [KrakenD](https://www.krakend.io) API gateway configuration file `krakend.json` and a Docker Compose file `docker-compose.yaml` for integration test.

## KrakenD

KrakenD is a binary executable processing the configuration file `krakend.json` for the API Gateway. 

The current used KrakenD version is `2.1.0` with Go `1.19.1` and Alpine `3.16`

### CI/CD

The latest images will be published to Docker Hub [repository](https://hub.docker.com/r/instill/api-gateway) at release.

## License

See the [LICENSE](./LICENSE) file for licensing information.
