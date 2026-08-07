# Azure DNS for `libdns`

This package implements the libdns interfaces for the [Azure DNS API](https://docs.microsoft.com/en-us/rest/api/dns/).

## Authenticating

This package supports authentication using **a service principal with a secret** and **a managed identity** through [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go).

### Service Principal with a Secret

To attempt to authenticate using a service principal with a secret, pass `TenantId`, `ClientId`, and `ClientSecret` to the `Provider`. If any of these three values are not empty, this package will attempt to authenticate using a service principal with a secret.

You will need to create a service principal using [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/create-an-azure-service-principal-azure-cli) or [Azure Portal](https://docs.microsoft.com/en-us/azure/active-directory/develop/howto-create-service-principal-portal), and assign the **DNS Zone Contributor** role to the service principal for the DNS zones that you want to manage.

Then keep the following information to pass to the `Provider` struct fields for authentication:

- `SubscriptionId` (`json:"subscription_id"`)
  - [DNS zones] > Your Zone > [Subscription ID]
- `ResourceGroupName` (`json:"resource_group_name"`)
  - [DNS zones] > Your Zone > [Resource group]
- `TenantId` (`json:"tenant_id"`)
  - [Microsoft Entra ID] > [Properties] > [Tenant ID]
- `ClientId` (`json:"client_id"`)
  - [Microsoft Entra ID] > [App registrations] > Your Application > [Application (client) ID]
- `ClientSecret` (`json:"client_secret"`)
  - [Microsoft Entra ID] > [App registrations] > Your Application > [Certificates & secrets] > [Client secrets] > [Value]

### Managed Identity

To attempt to authenticate using a managed identity, leave all of `TenantId`, `ClientId`, and `ClientSecret` unset or empty to the `Provider`. If all three values are unset or empty, this package will attempt to authenticate using a managed identity.

You will need to assign the **DNS Zone Contributor** role to the managed identity for the DNS zones that you want to manage.

Then keep the following information to pass to the `Provider` struct fields for authentication:

- `SubscriptionId` (`json:"subscription_id"`)
  - [DNS zones] > Your Zone > [Subscription ID]
- `ResourceGroupName` (`json:"resource_group_name"`)
  - [DNS zones] > Your Zone > [Resource group]

> [!NOTE]
> If this package is running outside of an Azure VM like Azure Arc, ensure required environment variables to use a managed identity (`IDENTITY_ENDPOINT`, `IMDS_ENDPOINT`, etc.) are available on your resources. [azure-sdk-for-go](https://github.com/Azure/azure-sdk-for-go) uses some environment variables to determine the endpoint for IMDS or HIMDS, and this package is also in the same manner. Refer to the Azure documentation for each services to use a managed identity.

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider (see `_example/main.go`)

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/libdns/azure"
	"github.com/libdns/libdns"
)

// main shows how libdns works with Azure DNS.
//
// In this example, the information required for authentication is passed as environment variables.
func main() {

	// Create new provider instance by authenticating using a service principal with a secret.
	// To authenticate using a managed identity, remove TenantId, ClientId, and ClientSecret.
	provider := azure.Provider{
		SubscriptionId:    os.Getenv("AZURE_SUBSCRIPTION_ID"),
		ResourceGroupName: os.Getenv("AZURE_RESOURCE_GROUP_NAME"),
		TenantId:          os.Getenv("AZURE_TENANT_ID"),
		ClientId:          os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:      os.Getenv("AZURE_CLIENT_SECRET"),
	}
	zone := os.Getenv("AZURE_DNS_ZONE_FQDN")

	// List existing records
	fmt.Printf("List existing records\n")
	currentRecords, err := provider.GetRecords(context.TODO(), zone)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	for _, record := range currentRecords {
		fmt.Printf("Exists: %v\n", record)
	}
}
```
