infomaniak for [`libdns`](https://github.com/libdns/libdns)
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/libdns/infomaniak)

This package implements the [libdns interfaces](https://github.com/libdns/libdns) for [infomaniak](https://infomaniak.com) based on their [API reference](https://developer.infomaniak.com/docs/api/get/2/zones/%7Bzone%7D/records), allowing you to manage DNS records.

## Code example
```go
import "github.com/libdns/infomaniak"
provider := &infomaniak.Provider{
    APIToken:  "YOUR_API_TOKEN"
}
```

## Create Your API Token
Please login to your infomaniak account and then navigate [here](https://manager.infomaniak.com/v3/infomaniak-api) to issue your API access token. The scope of your token has to include "domain:read", "dns:read" and "dns:write".
> :warning: All releases up to and including v0.1.3 use an unsupported API version and **require a different scope for your token**. If you use any of these releases, please make sure the "domain" scope is include for your token instead of the ones listed above.

## Development
### Setup
The repository contains configurations for a Visual Studio Code dev container. Please install the Visual Studio Code extension `ms-vscode-remote.remote-containers` to make use of it.

### Testing With Caddy
If you use the provided Dev Container, you can easily test your changes directly with caddy. Start by replacing the three placeholders for email, API token and domain in `.devcontainer/.caddyfile` (make sure to never commit these changes). You can then use the commands below directly in your Dev Container.

After each change, you have to rebuild caddy with the following command - make sure to enter a valid version of caddy:
`xcaddy build <caddy_version> --with dns.providers.infomaniak=/workspaces/caddy-dns-infomaniak --replace github.com/libdns/infomaniak=/workspaces/libdns-infomaniak --output /workspaces/caddy/caddy`

Run caddy with the following command and monitor the output:
`/workspaces/caddy/caddy run --config /workspaces/caddy/.caddyfile`

If caddy managed to successfully issue the certificate and you would like to test from scratch again, you have to delete the previously issued certificate by running
`rm -R /home/vscode/.local/share/caddy`
