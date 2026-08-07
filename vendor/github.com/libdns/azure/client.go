package azure

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"

	"github.com/libdns/libdns"
)

// Client is an abstraction of RecordSetsClient for Azure DNS
type Client struct {
	azureClient *armdns.RecordSetsClient
	mutex       sync.Mutex
}

// setupClient invokes authentication and store client to the provider instance.
func (p *Provider) setupClient() error {
	if p.client.azureClient == nil {
		credentials := []azcore.TokenCredential{}

		// If Tenant ID, Client ID, or Client Secret is specified, attempt to authenticate using a client secret.
		// If not, attempt to authenticate using managed identity.
		// Authentication using a client secret is prioritized over using managed identiry to keep backward compatibility.
		if p.TenantId != "" || p.ClientId != "" || p.ClientSecret != "" {
			clientCredential, err := azidentity.NewClientSecretCredential(p.TenantId, p.ClientId, p.ClientSecret, nil)
			if err != nil {
				return err
			}
			credentials = append(credentials, clientCredential)
		} else {
			managedIdentityCredential, err := azidentity.NewManagedIdentityCredential(nil)
			if err != nil {
				return err
			}
			credentials = append(credentials, managedIdentityCredential)
		}

		chainedTokenCredential, err := azidentity.NewChainedTokenCredential(credentials, nil)
		if err != nil {
			return err
		}
		clientFactory, err := armdns.NewClientFactory(p.SubscriptionId, chainedTokenCredential, nil)
		if err != nil {
			return err
		}
		p.client.azureClient = clientFactory.NewRecordSetsClient()
	}

	return nil
}

// getRecords gets all records in specified zone on Azure DNS.
func (p *Provider) getRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return nil, err
	}

	var recordSets []*armdns.RecordSet

	pager := p.client.azureClient.NewListByDNSZonePager(
		p.ResourceGroupName,
		strings.TrimSuffix(zone, "."),
		&armdns.RecordSetsClientListByDNSZoneOptions{
			Top:                 nil,
			Recordsetnamesuffix: nil,
		})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		recordSets = append(recordSets, page.Value...)
	}

	records, _ := convertAzureRecordSetsToLibdnsRecords(recordSets)
	return records, nil
}

// createRecord creates a new record in the specified zone.
// It throws an error if the record already exists.
func (p *Provider) createRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	return p.createOrUpdateRecord(ctx, zone, record, "*")
}

// updateRecord creates or updates a record, either by updating existing record or creating new one.
func (p *Provider) updateRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	return p.createOrUpdateRecord(ctx, zone, record, "")
}

// deleteRecord deletes an existing records.
// Regardless of the value of the record, if the name and type match, the record will be deleted.
func (p *Provider) deleteRecord(ctx context.Context, zone string, record libdns.Record) (libdns.Record, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return record, err
	}

	recordType, err := convertStringToRecordType(record.Type)
	if err != nil {
		return record, err
	}

	_, err = p.client.azureClient.Delete(
		ctx,
		p.ResourceGroupName,
		strings.TrimSuffix(zone, "."),
		generateRecordSetName(record.Name, zone),
		recordType,
		&armdns.RecordSetsClientDeleteOptions{
			IfMatch: nil,
		},
	)
	if err != nil {
		return record, err
	}

	return record, nil
}

// createOrUpdateRecord creates or updates a record.
// The behavior depends on the value of ifNoneMatch, set to "*" to allow to create a new record but prevent updating an existing record.
func (p *Provider) createOrUpdateRecord(ctx context.Context, zone string, record libdns.Record, ifNoneMatch string) (libdns.Record, error) {
	p.client.mutex.Lock()
	defer p.client.mutex.Unlock()

	if err := p.setupClient(); err != nil {
		return record, err
	}

	recordType, err := convertStringToRecordType(record.Type)
	if err != nil {
		return record, err
	}

	recordSet, err := convertLibdnsRecordToAzureRecordSet(record)
	if err != nil {
		return record, err
	}

	_, err = p.client.azureClient.CreateOrUpdate(
		ctx,
		p.ResourceGroupName,
		strings.TrimSuffix(zone, "."),
		generateRecordSetName(record.Name, zone),
		recordType,
		recordSet,
		&armdns.RecordSetsClientCreateOrUpdateOptions{
			IfMatch:     nil,
			IfNoneMatch: &ifNoneMatch,
		},
	)
	if err != nil {
		return record, err
	}

	return record, nil
}

// generateRecordSetName generates name for RecordSet object.
func generateRecordSetName(name string, zone string) string {
	recordSetName := libdns.RelativeName(strings.TrimSuffix(name, ".")+".", zone)
	if recordSetName == "" {
		return "@"
	}
	return recordSetName
}

// convertStringToRecordType casts standard type name string to an Azure-styled dedicated type.
func convertStringToRecordType(typeName string) (armdns.RecordType, error) {
	switch typeName {
	case "A":
		return armdns.RecordTypeA, nil
	case "AAAA":
		return armdns.RecordTypeAAAA, nil
	case "CAA":
		return armdns.RecordTypeCAA, nil
	case "CNAME":
		return armdns.RecordTypeCNAME, nil
	case "MX":
		return armdns.RecordTypeMX, nil
	case "NS":
		return armdns.RecordTypeNS, nil
	case "PTR":
		return armdns.RecordTypePTR, nil
	case "SOA":
		return armdns.RecordTypeSOA, nil
	case "SRV":
		return armdns.RecordTypeSRV, nil
	case "TXT":
		return armdns.RecordTypeTXT, nil
	default:
		return armdns.RecordTypeA, fmt.Errorf("the type %v cannot be interpreted", typeName)
	}
}

// convertAzureRecordSetsToLibdnsRecords converts Azure-styled records to libdns records.
func convertAzureRecordSetsToLibdnsRecords(recordSets []*armdns.RecordSet) ([]libdns.Record, error) {
	var records []libdns.Record

	for _, recordSet := range recordSets {
		switch typeName := strings.TrimPrefix(*recordSet.Type, "Microsoft.Network/dnszones/"); typeName {
		case "A":
			for _, v := range recordSet.Properties.ARecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: *v.IPv4Address,
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "AAAA":
			for _, v := range recordSet.Properties.AaaaRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: *v.IPv6Address,
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "CAA":
			for _, v := range recordSet.Properties.CaaRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: strings.Join([]string{fmt.Sprint(*v.Flags), *v.Tag, *v.Value}, " "),
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "CNAME":
			record := libdns.Record{
				ID:    *recordSet.Etag,
				Type:  typeName,
				Name:  *recordSet.Name,
				Value: *recordSet.Properties.CnameRecord.Cname,
				TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
			}
			records = append(records, record)
		case "MX":
			for _, v := range recordSet.Properties.MxRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: strings.Join([]string{fmt.Sprint(*v.Preference), *v.Exchange}, " "),
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "NS":
			for _, v := range recordSet.Properties.NsRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: *v.Nsdname,
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "PTR":
			for _, v := range recordSet.Properties.PtrRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: *v.Ptrdname,
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "SOA":
			record := libdns.Record{
				ID:   *recordSet.Etag,
				Type: typeName,
				Name: *recordSet.Name,
				Value: strings.Join([]string{
					*recordSet.Properties.SoaRecord.Host,
					*recordSet.Properties.SoaRecord.Email,
					fmt.Sprint(*recordSet.Properties.SoaRecord.SerialNumber),
					fmt.Sprint(*recordSet.Properties.SoaRecord.RefreshTime),
					fmt.Sprint(*recordSet.Properties.SoaRecord.RetryTime),
					fmt.Sprint(*recordSet.Properties.SoaRecord.ExpireTime),
					fmt.Sprint(*recordSet.Properties.SoaRecord.MinimumTTL)},
					" "),
				TTL: time.Duration(*recordSet.Properties.TTL) * time.Second,
			}
			records = append(records, record)
		case "SRV":
			for _, v := range recordSet.Properties.SrvRecords {
				record := libdns.Record{
					ID:    *recordSet.Etag,
					Type:  typeName,
					Name:  *recordSet.Name,
					Value: strings.Join([]string{fmt.Sprint(*v.Priority), fmt.Sprint(*v.Weight), fmt.Sprint(*v.Port), *v.Target}, " "),
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
				}
				records = append(records, record)
			}
		case "TXT":
			for _, v := range recordSet.Properties.TxtRecords {
				for _, txt := range v.Value {
					record := libdns.Record{
						ID:    *recordSet.Etag,
						Type:  typeName,
						Name:  *recordSet.Name,
						Value: *txt,
						TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
					}
					records = append(records, record)
				}
			}
		default:
			return []libdns.Record{}, fmt.Errorf("the type %v cannot be interpreted", typeName)
		}
	}

	return records, nil
}

// convertLibdnsRecordToAzureRecordSet converts a libdns record to an Azure-styled record.
func convertLibdnsRecordToAzureRecordSet(record libdns.Record) (armdns.RecordSet, error) {
	switch record.Type {
	case "A":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				ARecords: []*armdns.ARecord{{
					IPv4Address: to.Ptr(record.Value),
				}},
			},
		}
		return recordSet, nil
	case "AAAA":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				AaaaRecords: []*armdns.AaaaRecord{{
					IPv6Address: to.Ptr(record.Value),
				}},
			},
		}
		return recordSet, nil
	case "CAA":
		values := strings.Split(record.Value, " ")
		flags, _ := strconv.ParseInt(values[0], 10, 32)
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				CaaRecords: []*armdns.CaaRecord{{
					Flags: to.Ptr[int32](int32(flags)),
					Tag:   to.Ptr(values[1]),
					Value: to.Ptr(values[2]),
				}},
			},
		}
		return recordSet, nil
	case "CNAME":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				CnameRecord: &armdns.CnameRecord{
					Cname: to.Ptr(record.Value),
				},
			},
		}
		return recordSet, nil
	case "MX":
		values := strings.Split(record.Value, " ")
		preference, _ := strconv.ParseInt(values[0], 10, 32)
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				MxRecords: []*armdns.MxRecord{{
					Preference: to.Ptr[int32](int32(preference)),
					Exchange:   to.Ptr(values[1]),
				}},
			},
		}
		return recordSet, nil
	case "NS":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				NsRecords: []*armdns.NsRecord{{
					Nsdname: to.Ptr(record.Value),
				}},
			},
		}
		return recordSet, nil
	case "PTR":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				PtrRecords: []*armdns.PtrRecord{{
					Ptrdname: to.Ptr(record.Value),
				}},
			},
		}
		return recordSet, nil
	case "SOA":
		values := strings.Split(record.Value, " ")
		serialNumber, _ := strconv.ParseInt(values[2], 10, 64)
		refreshTime, _ := strconv.ParseInt(values[3], 10, 64)
		retryTime, _ := strconv.ParseInt(values[4], 10, 64)
		expireTime, _ := strconv.ParseInt(values[5], 10, 64)
		minimumTTL, _ := strconv.ParseInt(values[6], 10, 64)
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				SoaRecord: &armdns.SoaRecord{
					Host:         to.Ptr(values[0]),
					Email:        to.Ptr(values[1]),
					SerialNumber: to.Ptr[int64](serialNumber),
					RefreshTime:  to.Ptr[int64](refreshTime),
					RetryTime:    to.Ptr[int64](retryTime),
					ExpireTime:   to.Ptr[int64](expireTime),
					MinimumTTL:   to.Ptr[int64](minimumTTL),
				},
			},
		}
		return recordSet, nil
	case "SRV":
		values := strings.Split(record.Value, " ")
		priority, _ := strconv.ParseInt(values[0], 10, 32)
		weight, _ := strconv.ParseInt(values[1], 10, 32)
		port, _ := strconv.ParseInt(values[2], 10, 32)
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				SrvRecords: []*armdns.SrvRecord{{
					Priority: to.Ptr[int32](int32(priority)),
					Weight:   to.Ptr[int32](int32(weight)),
					Port:     to.Ptr[int32](int32(port)),
					Target:   to.Ptr(values[3]),
				}},
			},
		}
		return recordSet, nil
	case "TXT":
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr[int64](int64(record.TTL / time.Second)),
				TxtRecords: []*armdns.TxtRecord{{
					Value: []*string{&record.Value},
				}},
			},
		}
		return recordSet, nil
	default:
		return armdns.RecordSet{}, fmt.Errorf("the type %v cannot be interpreted", record.Type)
	}
}
