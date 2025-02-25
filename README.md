# Who Did The Chores

The purpose of this application is to help households keep track on who did what concerning chores.
It lets you define users, chores and compile when a user do a chore (named task in the application).

The completed tasks can then be visualized in a graph.

## Quickstart

The application is packaged in a [Docker image](https://hub.docker.com/repository/docker/mqufflc/whodidthechores).

A docker compose file and a helm chart are available to deploy the application.

### Docker Compose

You will need [git](https://git-scm.com/downloads), [docker](https://docs.docker.com/engine/install/) and [docker compose](https://docs.docker.com/compose/install/linux/) plugin to use the following commmands:

```bash
git clone https://github.com/mqufflc/whodidthechores.git
cd whodidthechores
docker compose up -d
```

### Helm

You will need [git](https://git-scm.com/downloads), [helm](https://helm.sh/docs/intro/install/) and a running [kubernetes](https://kubernetes.io) cluster to use the following commands:

The helm chart is not yet available in a repository, you can still use the following commands in the meantime to use it:

```bash
git clone https://github.com/mqufflc/whodidthechores.git
cd whodidthechores
helm install whodidthechores helm/whodidthechores/
```

## Disclaimer

This project is working but a lot of work is still needed. If you want to use it, you will definitely encounter bugs.