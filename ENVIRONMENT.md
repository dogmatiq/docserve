# Environment Variables

This document describes the environment variables used by `browser`.

⚠️ Some of the variables have **non-normative** examples. These examples are
syntactically correct but may not be meaningful values for this application.

⚠️ The application may consume other undocumented environment variables; this
document only shows those variables declared using [Ferrite].

Please note that **undefined** variables and **empty strings** are considered
equivalent.

## Index

- [`DSN`](#DSN) — the PostgreSQL connection string
- [`GITHUB_APP_CLIENT_ID`](#GITHUB_APP_CLIENT_ID) — the client ID of the GitHub application used to read repository content
- [`GITHUB_APP_CLIENT_SECRET`](#GITHUB_APP_CLIENT_SECRET) — the client secret for the GitHub application used to read repository content
- [`GITHUB_APP_ID`](#GITHUB_APP_ID) — the ID of the GitHub application used to read repository content
- [`GITHUB_APP_PRIVATE_KEY`](#GITHUB_APP_PRIVATE_KEY) — the private key for the GitHub application used to read repository content
- [`GITHUB_HOOK_SECRET`](#GITHUB_HOOK_SECRET) — the secret used to verify GitHub web-hook requests are genuine
- [`GITHUB_URL`](#GITHUB_URL) — the base URL of the GitHub API

## Specification

### `DSN`

> the PostgreSQL connection string

This variable **MUST** be set to a non-empty string.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export DSN=foo # (non-normative)
```

### `GITHUB_APP_CLIENT_ID`

> the client ID of the GitHub application used to read repository content

This variable **MUST** be set to a non-empty string.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export GITHUB_APP_CLIENT_ID=foo # (non-normative)
```

### `GITHUB_APP_CLIENT_SECRET`

> the client secret for the GitHub application used to read repository content

This variable **MUST** be set to a non-empty string.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export GITHUB_APP_CLIENT_SECRET=foo # (non-normative)
```

### `GITHUB_APP_ID`

> the ID of the GitHub application used to read repository content

This variable **MUST** be set to `+1` or greater.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export GITHUB_APP_ID=+2305843009213693952 # (non-normative)
```

### `GITHUB_APP_PRIVATE_KEY`

> the private key for the GitHub application used to read repository content

This variable **MUST** be set to a non-empty string.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export GITHUB_APP_PRIVATE_KEY=foo # (non-normative)
```

### `GITHUB_HOOK_SECRET`

> the secret used to verify GitHub web-hook requests are genuine

This variable **MUST** be set to a non-empty string.
If left undefined the application will print usage information to `STDERR` then
exit with a non-zero exit code.

```bash
export GITHUB_HOOK_SECRET=foo # (non-normative)
```

### `GITHUB_URL`

> the base URL of the GitHub API

This variable **MAY** be set to a non-empty value or left undefined.

```bash
export GITHUB_URL=https://example.org/path # (non-normative) a typical URL for a web page
```

## Usage Examples

<details>
<summary>Kubernetes</summary>

This example shows how to define the environment variables needed by `browser`
on a [Kubernetes container] within a Kubenetes deployment manifest.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  template:
    spec:
      containers:
        - name: example-container
          env:
            - name: DSN # the PostgreSQL connection string
              value: foo
            - name: GITHUB_APP_CLIENT_ID # the client ID of the GitHub application used to read repository content
              value: foo
            - name: GITHUB_APP_CLIENT_SECRET # the client secret for the GitHub application used to read repository content
              value: foo
            - name: GITHUB_APP_ID # the ID of the GitHub application used to read repository content
              value: "+2305843009213693952"
            - name: GITHUB_APP_PRIVATE_KEY # the private key for the GitHub application used to read repository content
              value: foo
            - name: GITHUB_HOOK_SECRET # the secret used to verify GitHub web-hook requests are genuine
              value: foo
            - name: GITHUB_URL # the base URL of the GitHub API
              value: https://example.org/path
```

Alternatively, the environment variables can be defined within a [config map][kubernetes config map]
then referenced a deployment manifest using `configMapRef`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config-map
data:
  DSN: foo # the PostgreSQL connection string
  GITHUB_APP_CLIENT_ID: foo # the client ID of the GitHub application used to read repository content
  GITHUB_APP_CLIENT_SECRET: foo # the client secret for the GitHub application used to read repository content
  GITHUB_APP_ID: "+2305843009213693952" # the ID of the GitHub application used to read repository content
  GITHUB_APP_PRIVATE_KEY: foo # the private key for the GitHub application used to read repository content
  GITHUB_HOOK_SECRET: foo # the secret used to verify GitHub web-hook requests are genuine
  GITHUB_URL: https://example.org/path # the base URL of the GitHub API
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-deployment
spec:
  template:
    spec:
      containers:
        - name: example-container
          envFrom:
            - configMapRef:
                name: example-config-map
```

</details>

<details>
<summary>Docker</summary>

This example shows how to define the environment variables needed by `browser`
when running as a [Docker service] defined in a Docker compose file.

```yaml
service:
  example-service:
    environment:
      DSN: foo # the PostgreSQL connection string
      GITHUB_APP_CLIENT_ID: foo # the client ID of the GitHub application used to read repository content
      GITHUB_APP_CLIENT_SECRET: foo # the client secret for the GitHub application used to read repository content
      GITHUB_APP_ID: "+2305843009213693952" # the ID of the GitHub application used to read repository content
      GITHUB_APP_PRIVATE_KEY: foo # the private key for the GitHub application used to read repository content
      GITHUB_HOOK_SECRET: foo # the secret used to verify GitHub web-hook requests are genuine
      GITHUB_URL: https://example.org/path # the base URL of the GitHub API
```

</details>

<!-- references -->

[docker service]: https://docs.docker.com/compose/environment-variables/#set-environment-variables-in-containers
[ferrite]: https://github.com/dogmatiq/ferrite
[kubernetes config map]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables
[kubernetes container]: https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/#define-an-environment-variable-for-a-container
