DNS Made Easy API client
=======================

[![Go Reference](https://pkg.go.dev/badge/test.svg)](https://pkg.go.dev/github.com/john-k/dnsmadeeasy
)

A Golang client for [DNS Made Easy](https://dnsmadeeasy.com) against their [APIv2 endpoints](https://api-docs.dnsmadeeasy.com/)

# Testing
Create a `.env` file containing the varibles `DME_API_TOKEN` and `DME_API_SECRET` using credentials from your DNS Made Easy Sandbox account, then run `go test -v`

NOTE: depending on the load on the DNS Made Easy sandbox environment, it may take an inordinate amount of time to finish creating the two domains that are created during testing. If either of the `TestSandboxIntegration/Cleanup_test_domains` calls fail, it will be necessary to manually delete those domains.