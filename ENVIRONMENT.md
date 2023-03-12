# Environment Variables

This document describes the environment variables used by `browser`.

If any of the environment variable values do not meet the requirements herein,
the application will print usage information to `STDERR` then exit with a
non-zero exit code. Please note that **undefined** variables and **empty**
values are considered equivalent.

⚠️ This document includes **non-normative** example values. While these values
are syntactically correct, they may not be meaningful to this application.

⚠️ The application may consume other undocumented environment variables; this
document only shows those variables declared using [Ferrite].

The key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**,
**SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY**, and **OPTIONAL** in this
document are to be interpreted as described in [RFC 2119].

## Index

- [`DSN`] — the PostgreSQL connection string
- [`GITHUB_APP_ID`] — the ID of the GitHub application used to read repository content
- [`GITHUB_APP_PRIVATEKEY`] — the private key for the GitHub application used to read repository content
- [`GITHUB_CLIENT_ID`] — the client ID of the GitHub application used to read repository content
- [`GITHUB_CLIENT_SECRET`] — the client secret for the GitHub application used to read repository content
- [`GITHUB_HOOK_SECRET`] — the secret used to verify GitHub web-hook requests are genuine
- [`GITHUB_URL`] — the base URL of the GitHub API

## Specification

### `DSN`

> the PostgreSQL connection string

The `DSN` variable **MUST NOT** be left undefined.

```bash
export DSN=foo # (non-normative)
```

### `GITHUB_APP_ID`

> the ID of the GitHub application used to read repository content

The `GITHUB_APP_ID` variable's value **MUST** be `1` or greater.

```bash
export GITHUB_APP_ID=1                    # (non-normative) the minimum accepted value
export GITHUB_APP_ID=8301034833169298432  # (non-normative)
export GITHUB_APP_ID=11068046444225730560 # (non-normative)
```

<details>
<summary>Unsigned integer syntax</summary>

Unsigned integers can only be specified using decimal (base-10) notation. A
leading sign (`+` or `-`) is not supported and **MUST NOT** be specified.

Internally, the `GITHUB_APP_ID` variable is represented using an unsigned 64-bit
integer type (`uint`); any value that overflows this data-type is invalid.

</details>

### `GITHUB_APP_PRIVATEKEY`

> the private key for the GitHub application used to read repository content

The `GITHUB_APP_PRIVATEKEY` variable **MUST NOT** be left undefined.

```bash
export GITHUB_APP_PRIVATEKEY=foo # (non-normative)
```

### `GITHUB_CLIENT_ID`

> the client ID of the GitHub application used to read repository content

The `GITHUB_CLIENT_ID` variable **MUST NOT** be left undefined.

```bash
export GITHUB_CLIENT_ID=foo # (non-normative)
```

### `GITHUB_CLIENT_SECRET`

> the client secret for the GitHub application used to read repository content

The `GITHUB_CLIENT_SECRET` variable **MUST NOT** be left undefined.

```bash
export GITHUB_CLIENT_SECRET=foo # (non-normative)
```

### `GITHUB_HOOK_SECRET`

> the secret used to verify GitHub web-hook requests are genuine

The `GITHUB_HOOK_SECRET` variable **MUST NOT** be left undefined.

```bash
export GITHUB_HOOK_SECRET=foo # (non-normative)
```

### `GITHUB_URL`

> the base URL of the GitHub API

The `GITHUB_URL` variable **MAY** be left undefined. Otherwise, the value
**MUST** be a fully-qualified URL.

```bash
export GITHUB_URL=https://example.org/path # (non-normative) a typical URL for a web page
```

<details>
<summary>URL syntax</summary>

A fully-qualified URL includes both a scheme (protocol) and a hostname. URLs are
not necessarily web addresses; `https://example.org` and
`mailto:contact@example.org` are both examples of fully-qualified URLs.

</details>

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
            - name: GITHUB_APP_ID # the ID of the GitHub application used to read repository content
              value: "1"
            - name: GITHUB_APP_PRIVATEKEY # the private key for the GitHub application used to read repository content
              value: foo
            - name: GITHUB_CLIENT_ID # the client ID of the GitHub application used to read repository content
              value: foo
            - name: GITHUB_CLIENT_SECRET # the client secret for the GitHub application used to read repository content
              value: foo
            - name: GITHUB_HOOK_SECRET # the secret used to verify GitHub web-hook requests are genuine
              value: foo
            - name: GITHUB_URL # the base URL of the GitHub API (optional)
              value: https://example.org/path
```

Alternatively, the environment variables can be defined within a [config map][kubernetes config map]
then referenced from a deployment manifest using `configMapRef`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config-map
data:
  DSN: foo # the PostgreSQL connection string
  GITHUB_APP_ID: "1" # the ID of the GitHub application used to read repository content
  GITHUB_APP_PRIVATEKEY: foo # the private key for the GitHub application used to read repository content
  GITHUB_CLIENT_ID: foo # the client ID of the GitHub application used to read repository content
  GITHUB_CLIENT_SECRET: foo # the client secret for the GitHub application used to read repository content
  GITHUB_HOOK_SECRET: foo # the secret used to verify GitHub web-hook requests are genuine
  GITHUB_URL: https://example.org/path # the base URL of the GitHub API (optional)
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
      GITHUB_APP_ID: "1" # the ID of the GitHub application used to read repository content
      GITHUB_APP_PRIVATEKEY: foo # the private key for the GitHub application used to read repository content
      GITHUB_CLIENT_ID: foo # the client ID of the GitHub application used to read repository content
      GITHUB_CLIENT_SECRET: foo # the client secret for the GitHub application used to read repository content
      GITHUB_HOOK_SECRET: foo # the secret used to verify GitHub web-hook requests are genuine
      GITHUB_URL: https://example.org/path # the base URL of the GitHub API (optional)
```

</details>

<!-- references -->

[docker service]: https://docs.docker.com/compose/environment-variables/#set-environment-variables-in-containers
[`dsn`]: #DSN
[ferrite]: https://github.com/dogmatiq/ferrite
[`github_app_id`]: #GITHUB_APP_ID
[`github_app_privatekey`]: #GITHUB_APP_PRIVATEKEY
[`github_client_id`]: #GITHUB_CLIENT_ID
[`github_client_secret`]: #GITHUB_CLIENT_SECRET
[`github_hook_secret`]: #GITHUB_HOOK_SECRET
[`github_url`]: #GITHUB_URL
[kubernetes config map]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#configure-all-key-value-pairs-in-a-configmap-as-container-environment-variables
[kubernetes container]: https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/#define-an-environment-variable-for-a-container
[rfc 2119]: https://www.rfc-editor.org/rfc/rfc2119.html
