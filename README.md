# API Gateway <!-- omit in toc -->

This repository maintains the [KrakenD](https://www.krakend.io) API gateway configuration file `krakend.json` and a Docker Compose file `docker-compose.yaml` for integration test.

- [How to use it](#how-to-use-it)
  - [KrakenD Designer](#krakend-designer)
  - [Reference](#reference)
- [Development and production environments](#development-and-production-environments)
- [Docker compose suite](#docker-compose-suite)
- [Integration tests](#integration-tests)
- [CI/CD](#cicd)
  - [Push](#push)
  - [Pull request](#pull-request)
  - [Image purge](#image-purge)

## How to use it

KrakenD is a binary executable processing the configuration file `krakend.json` for operating the API Gateway. The `Dockerfile` in this repository uses `devopsfaith/krakend:1.4.1` as the base image.

We use [elasticsearch-certutil](https://www.elastic.co/guide/en/elasticsearch/reference/current/certutil.html) for generating TLS certificate files shared with the API Gateway and other backend components. We also leverage [elasticsearch-keystore](https://www.elastic.co/guide/en/elasticsearch/reference/current/elasticsearch-keystore.html) for storing secrets.

### KrakenD Designer

[KrakenD Designer](https://github.com/devopsfaith/api-gatewayesigner) is ideal for configuring the API Gateway via a Web UI:
```bash
docker run --rm -p 8081:80 devopsfaith/api-gatewayesigner
```

### Reference
- [Best practices](https://www.krakend.io/docs/deploying/best-practices)

## Development and production environments
Docker images are provided for
- **development** (`api-gateway/Dockerfile` - target `dev`): this Docker image contains configuration for local development environment for the API gateway. This image is used in the [docker compose suite](#docker-compose-suite) and should not be pushed to the Harbor container repository.
- **production** (`api-gateway/Dockerfile` - target `prod`): this Docker image contains configuration for Kubernetes environment for production usage and will be pushed to Harbor container repository.

## Docker compose suite

The `docker-compose.*.yml` provides an integration setup for [`api-gateway`](https://github.com/instill-ai/api-gateway) ⟷ [`model-backend`](https://github.com/instill-ai/model-backend) ⟷ [`triton-backend`](https://github.com/instill-ai/triton-backend), together with monitoring and logging stacks for local development:

See what docker compose functions are available:
```bash
make or make help
```

To launch all Instill services:
```bash
make dev
```

Now you can query the endpoints like:
```bash
curl -k --location --request POST 'https://localhost:8000/demo/tasks/{classification,detection}/outputs' \
--header 'Content-Type: application/json' \
--data-raw '{
    "contents": [
        {
            "url": <image url>
        }
    ]
}'
```

To launch monitoring stack:
```bash
make monitoring
```

To launch logging stack:
```bash
make logging
```

To launch all services and stacks:
```bash
make all
```

To prune everything:
```bash
make prune
```

## Integration tests
Make sure all backend containers are running via `make dev`. Follow the guideline to install [k6](https://k6.io/docs/getting-started/installation) and run
```
make test
```

Our k6 test relies on a k6 extension [xk6-jose](https://github.com/szkiba/xk6-jose#build). If you want to run k6 test manually:
1. Install `xk6`
  ```bash
  $ go install go.k6.io/xk6/cmd/xk6@latest
  ```

2. Build the binary
  ```bash
  $ xk6 build --with github.com/szkiba/xk6-jose@latest
  ```
3. Use the built binary `k6` to run the tests

## CI/CD

### Push
With [Release Please Action](https://github.com/google-github-actions/release-please-action), we maintain two types of version for the container images, (almost) following [SemVer 2.0](https://semver.org):
1. Version core for release `push`: `<versioncore>`;
2. Build metadata version for non-release `push`: `<versioncore>_<buildmetadata>`

Unlike what is specified in [SemVer 2.0](https://semver.org), we use `_` instead of `+` to separate the version core and the build metadata because [Docker tags don't support `+`](https://github.com/opencontainers/distribution-spec/issues/154).

### Pull request
Each git `push` to a feature branch in a pull request (PR) session will trigger a container build tagged by its build metadata version and pushed to Harbor image repository.

### Image purge
Two cases:
1. Images with a build metadata version will be purged every time after a PR is merged (pushed) into `main` branch. The image with the latest build metadata version (latest commit) will be kept.

2. All images with a build metadata version will be purged right after a new release is issued.
