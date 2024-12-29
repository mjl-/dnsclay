package azure

import (
	"context"

	"github.com/libdns/libdns"
)

// Provider implements the libdns interfaces for Azure DNS
type Provider struct {

	// Subscription ID is the ID of the subscription in which the DNS zone is located. Required.
	SubscriptionId string `json:"subscription_id,omitempty"`

	// Resource Group Name is the name of the resource group in which the DNS zone is located. Required.
	ResourceGroupName string `json:"resource_group_name,omitempty"`

	// (Optional)
	// Tenant ID is the ID of the tenant of the Microsoft Entra ID in which the application is located.
	// Required only when authenticating using a service principal with a secret.
	// Do not set any value to authenticate using a managed identity.
	TenantId string `json:"tenant_id,omitempty"`

	// (Optional)
	// Client ID is the ID of the application.
	// Required only when authenticating using a service principal with a secret.
	// Do not set any value to authenticate using a managed identity.
	ClientId string `json:"client_id,omitempty"`

	// (Optional)
	// Client Secret is the client secret of the application.
	// Required only when authenticating using a service principal with a secret.
	// Do not set any value to authenticate using a managed identity.
	ClientSecret string `json:"client_secret,omitempty"`

	client Client
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	records, err := p.getRecords(ctx, zone)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var createdRecords []libdns.Record

	for _, record := range records {
		createdRecord, err := p.createRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		createdRecords = append(createdRecords, createdRecord)
	}

	return createdRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var updatedRecords []libdns.Record

	for _, record := range records {
		updatedRecord, err := p.updateRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		updatedRecords = append(updatedRecords, updatedRecord)
	}

	return updatedRecords, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	for _, record := range records {
		deletedRecord, err := p.deleteRecord(ctx, zone, record)
		if err != nil {
			return nil, err
		}
		deletedRecords = append(deletedRecords, deletedRecord)
	}

	return deletedRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
