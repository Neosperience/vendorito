# Vendorito

Vendorito is a CLI tool to copy OCI images across registries. The main purpose of this tool is being able to vendor images from third parties in owned repositories avoiding the need for a registry proxy, eg. copying a Docker Hub image to ECR/Quay to avoid Docker Hub pull limits.

In essence, this tool is basically the `copy` command of [skopeo] with a more straightforward CLI and the ability to specify docker credentials in multiple ways (Docker config file, inline credentials, etc).

## Usage

The tool can be invoked with the following syntax:

```
$ vendorito -i <image-url> -o <target-url>
```

Credentials can be specified in multiple ways:
 - Inline credentials in the image URL, like this: `username:password@quay.io/my/image`. Like inline basic auth URLs, the password has to be URL encoded.
 - Docker config file with auth credentials, specified via `-auth-file` CLI param or `VENDORITO_AUTH_FILE` environment variable.
 - Auth credentials in the format `domain.tld:username:password` in the `-credentials` CLI param or `VENDORITO_CREDENTIALS` environment variable. When using this format, multiple credentials can be specified, either by using multiple `-credentials` params or by separating them with a space when using the environment variable.

[skopeo]: https://github.com/containers/skopeo