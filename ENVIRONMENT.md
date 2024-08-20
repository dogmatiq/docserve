# Environment Variables

This document describes the environment variables used by `browser`.

| Name                      | Usage    | Description                                                                  |
| ------------------------- | -------- | ---------------------------------------------------------------------------- |
| [`DSN`]                   | required | the PostgreSQL connection string                                             |
| [`GITHUB_APP_ID`]         | required | the ID of the GitHub application used to read repository content             |
| [`GITHUB_APP_PRIVATEKEY`] | required | the private key for the GitHub application used to read repository content   |
| [`GITHUB_CLIENT_ID`]      | required | the client ID of the GitHub application used to read repository content      |
| [`GITHUB_CLIENT_SECRET`]  | required | the client secret for the GitHub application used to read repository content |
| [`GITHUB_HOOK_SECRET`]    | required | the secret used to verify GitHub web-hook requests are genuine               |
| [`GITHUB_URL`]            | optional | the base URL of the GitHub API                                               |

> [!TIP]
> If an environment variable is set to an empty value, `browser` behaves as if
> that variable is left undefined.

## `DSN`

> the PostgreSQL connection string

The `DSN` variable **MUST NOT** be left undefined.

```bash
export DSN=foo # (non-normative)
```

## `GITHUB_APP_ID`

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

## `GITHUB_APP_PRIVATEKEY`

> the private key for the GitHub application used to read repository content

The `GITHUB_APP_PRIVATEKEY` variable **MUST NOT** be left undefined.

⚠️ This variable is **sensitive**; its value may contain private information.

## `GITHUB_CLIENT_ID`

> the client ID of the GitHub application used to read repository content

The `GITHUB_CLIENT_ID` variable **MUST NOT** be left undefined.

```bash
export GITHUB_CLIENT_ID=foo # (non-normative)
```

## `GITHUB_CLIENT_SECRET`

> the client secret for the GitHub application used to read repository content

The `GITHUB_CLIENT_SECRET` variable **MUST NOT** be left undefined.

⚠️ This variable is **sensitive**; its value may contain private information.

## `GITHUB_HOOK_SECRET`

> the secret used to verify GitHub web-hook requests are genuine

The `GITHUB_HOOK_SECRET` variable **MUST NOT** be left undefined.

⚠️ This variable is **sensitive**; its value may contain private information.

## `GITHUB_URL`

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

---

> [!NOTE]
> This document only describes environment variables declared using [Ferrite].
> `browser` may consume other undocumented environment variables.

> [!IMPORTANT]
> Some of the example values given in this document are **non-normative**.
> Although these values are syntactically valid, they may not be meaningful to
> `browser`.

<!-- references -->

[`dsn`]: #DSN
[ferrite]: https://github.com/dogmatiq/ferrite
[`github_app_id`]: #GITHUB_APP_ID
[`github_app_privatekey`]: #GITHUB_APP_PRIVATEKEY
[`github_client_id`]: #GITHUB_CLIENT_ID
[`github_client_secret`]: #GITHUB_CLIENT_SECRET
[`github_hook_secret`]: #GITHUB_HOOK_SECRET
[`github_url`]: #GITHUB_URL
