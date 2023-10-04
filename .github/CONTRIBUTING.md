# Contributing Guidelines

We appreciate your contribution to this amazing project! Any form of engagement is welcome, including but not limiting to

- feature request
- documentation wording
- bug report
- roadmap suggestion
- ...and so on!

Please refer to the [community contributing section](https://github.com/instill-ai/community#contributing) for more details.

## Development and codebase contribution

Before delving into the details to come up with your first PR, please familiarise yourself with the project structure of [Instill Core](https://github.com/instill-ai/community#instill-core).

### Prerequisites

- [Instill Base](https://github.com/instill-ai/base)

### Local development

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
$ cd grpc_proxy_plugin && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugin/grpc-proxy.so /api-gateway/grpc_proxy_plugin/pkg && cd /api-gateway # compile the KrakenD grpc-proxy plugin
$ cd multi_auth_plugin && go build -buildmode=plugin -buildvcs=false -o /usr/local/lib/krakend/plugin/multi-auth.so /api-gateway/multi_auth_plugin/server && cd /api-gateway # compile the KrakenD multi-auth plugin
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

### Sending PRs

Please take these general guidelines into consideration when you are sending a PR:

1. **Fork the Repository:** Begin by forking the repository to your GitHub account.
2. **Create a New Branch:** Create a new branch to house your work. Use a clear and descriptive name, like `<your-github-username>/<what-your-pr-about>`.
3. **Make and Commit Changes:** Implement your changes and commit them. We encourage you to follow these best practices for commits to ensure an efficient review process:
   - Adhere to the [conventional commits guidelines](https://www.conventionalcommits.org/) for meaningful commit messages.
   - Follow the [7 rules of commit messages](https://chris.beams.io/posts/git-commit/) for well-structured and informative commits.
   - Rearrange commits to squash trivial changes together, if possible. Utilize [git rebase](http://gitready.com/advanced/2009/03/20/reorder-commits-with-rebase.html) for this purpose.
4. **Push to Your Branch:** Push your branch to your GitHub repository: `git push origin feat/<your-feature-name>`.
5. **Open a Pull Request:** Initiate a pull request to our repository. Our team will review your changes and collaborate with you on any necessary refinements.

When you are ready to send a PR, we recommend you to first open a `draft` one. This will trigger a bunch of `tests` [workflows](https://github.com/instill-ai/mgmt-backend/tree/main/.github/workflows) running a thorough test suite on multiple platforms. After the tests are done and passed, you can now mark the PR `open` to notify the codebase owners to review. We appreciate your endeavour to pass the integration test for your PR to make sure the sanity with respect to the entire scope of **Instill Core**.

## Last words

Your contributions make a difference. Let's build something amazing together!
