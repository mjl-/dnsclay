# TencentCloud DNSPod for `libdns`

This package implements the [libdns](https://github.com/libdns/libdns) interfaces for the [TencentCloud DNSPod API](https://www.tencentcloud.com/zh/document/api/1157/49025)

## Code example

```go
import "github.com/libdns/tencentcloud"
provider := &tencentcloud.Provider{
    SecretId:  "YOUR_Secret_ID",
    SecretKey: "YOUR_Secret_Key",
}
```

## Security Credentials

To authenticate you need to supply a [TencentCloud API Key](https://console.tencentcloud.com/cam/capi).

## Other instructions

`libdns/tencentcloud` is based on the new version of Tencentcloud api, uses secret Id and key as authentication methods, supports permission settings, and supports DNSPod international version.

`libdns/dnspod` is based on the old version of dnspod.cn api, uses token as the authentication method, does not support permission settings, and does not support DNSPod international version.
