package azure

import (
	"context"
	"fmt"
	"net/netip"
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

	rr := record.RR()
	recordType, err := convertStringToRecordType(rr.Type)
	if err != nil {
		return record, err
	}

	_, err = p.client.azureClient.Delete(
		ctx,
		p.ResourceGroupName,
		strings.TrimSuffix(zone, "."),
		generateRecordSetName(rr.Name, zone),
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

	rr := record.RR()
	recordType, err := convertStringToRecordType(rr.Type)
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
		generateRecordSetName(rr.Name, zone),
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
	case "SRV":
		return armdns.RecordTypeSRV, nil
	case "TXT":
		return armdns.RecordTypeTXT, nil
	case "PTR":
		return armdns.RecordTypePTR, nil
	case "SOA":
		return armdns.RecordTypeSOA, nil
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
				ip, err := netip.ParseAddr(*v.IPv4Address)
				if err != nil {
					return nil, fmt.Errorf("failed to parse IP address: %w", err)
				}
				record := libdns.Address{
					Name: *recordSet.Name,
					TTL:  time.Duration(*recordSet.Properties.TTL) * time.Second,
					IP:   ip,
				}
				records = append(records, record)
			}
		case "AAAA":
			for _, v := range recordSet.Properties.AaaaRecords {
				ip, err := netip.ParseAddr(*v.IPv6Address)
				if err != nil {
					return nil, fmt.Errorf("failed to parse IP address: %w", err)
				}
				record := libdns.Address{
					Name: *recordSet.Name,
					TTL:  time.Duration(*recordSet.Properties.TTL) * time.Second,
					IP:   ip,
				}
				records = append(records, record)
			}
		case "CAA":
			for _, v := range recordSet.Properties.CaaRecords {
				record := libdns.CAA{
					Name:  *recordSet.Name,
					TTL:   time.Duration(*recordSet.Properties.TTL) * time.Second,
					Flags: uint8(*v.Flags),
					Tag:   *v.Tag,
					Value: *v.Value,
				}
				records = append(records, record)
			}
		case "CNAME":
			record := libdns.CNAME{
				Name:   *recordSet.Name,
				TTL:    time.Duration(*recordSet.Properties.TTL) * time.Second,
				Target: *recordSet.Properties.CnameRecord.Cname,
			}
			records = append(records, record)
		case "MX":
			for _, v := range recordSet.Properties.MxRecords {
				record := libdns.MX{
					Name:       *recordSet.Name,
					TTL:        time.Duration(*recordSet.Properties.TTL) * time.Second,
					Preference: uint16(*v.Preference),
					Target:     *v.Exchange,
				}
				records = append(records, record)
			}
		case "NS":
			for _, v := range recordSet.Properties.NsRecords {
				record := libdns.NS{
					Name:   *recordSet.Name,
					TTL:    time.Duration(*recordSet.Properties.TTL) * time.Second,
					Target: *v.Nsdname,
				}
				records = append(records, record)
			}
		case "SRV":
			for _, v := range recordSet.Properties.SrvRecords {
				parts := strings.SplitN(*recordSet.Name, ".", 3)
				if len(parts) < 2 {
					return nil, fmt.Errorf("name %v does not contain enough fields; expected format: '_service._proto.name' or '_service._proto'", *recordSet.Name)
				}
				name := "@"
				if len(parts) == 3 {
					name = parts[2]
				}
				record := libdns.SRV{
					Service:   strings.TrimPrefix(parts[0], "_"),
					Transport: strings.TrimPrefix(parts[1], "_"),
					Name:      name,
					TTL:       time.Duration(*recordSet.Properties.TTL) * time.Second,
					Priority:  uint16(*v.Priority),
					Weight:    uint16(*v.Weight),
					Port:      uint16(*v.Port),
					Target:    *v.Target,
				}
				records = append(records, record)
			}
		case "TXT":
			for _, v := range recordSet.Properties.TxtRecords {
				for _, txt := range v.Value {
					record := libdns.TXT{
						Name: *recordSet.Name,
						TTL:  time.Duration(*recordSet.Properties.TTL) * time.Second,
						Text: *txt,
					}
					records = append(records, record)
				}
			}
		case "PTR":
			for _, v := range recordSet.Properties.PtrRecords {
				record := libdns.RR{
					Name: *recordSet.Name,
					Type: "PTR",
					TTL:  time.Duration(*recordSet.Properties.TTL) * time.Second,
					Data: *v.Ptrdname,
				}
				records = append(records, record)
			}
		case "SOA":
			soaData := strings.Join([]string{
				*recordSet.Properties.SoaRecord.Host,
				*recordSet.Properties.SoaRecord.Email,
				fmt.Sprint(*recordSet.Properties.SoaRecord.SerialNumber),
				fmt.Sprint(*recordSet.Properties.SoaRecord.RefreshTime),
				fmt.Sprint(*recordSet.Properties.SoaRecord.RetryTime),
				fmt.Sprint(*recordSet.Properties.SoaRecord.ExpireTime),
				fmt.Sprint(*recordSet.Properties.SoaRecord.MinimumTTL)},
				" ")
			record := libdns.RR{
				Name: *recordSet.Name,
				Type: "SOA",
				TTL:  time.Duration(*recordSet.Properties.TTL) * time.Second,
				Data: soaData,
			}
			records = append(records, record)
		default:
			return []libdns.Record{}, fmt.Errorf("the type %v cannot be interpreted", typeName)
		}
	}

	return records, nil
}

// convertLibdnsRecordToAzureRecordSet converts a libdns record to an Azure-styled record.
func convertLibdnsRecordToAzureRecordSet(record libdns.Record) (armdns.RecordSet, error) {
	fmt.Println("record:", record)
	rr, err := record.RR().Parse()
	if err != nil {
		return armdns.RecordSet{}, fmt.Errorf("unable to parse libdns.RR: %w", err)
	}

	switch rec := rr.(type) {
	case libdns.Address:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
			},
		}
		if rec.IP.Is6() {
			recordSet.Properties.AaaaRecords = []*armdns.AaaaRecord{{
				IPv6Address: to.Ptr(rec.IP.String()),
			}}
		} else {
			recordSet.Properties.ARecords = []*armdns.ARecord{{
				IPv4Address: to.Ptr(rec.IP.String()),
			}}
		}
		return recordSet, nil
	case libdns.CAA:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				CaaRecords: []*armdns.CaaRecord{{
					Flags: to.Ptr(int32(rec.Flags)),
					Tag:   to.Ptr(rec.Tag),
					Value: to.Ptr(rec.Value),
				}},
			},
		}
		return recordSet, nil
	case libdns.CNAME:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				CnameRecord: &armdns.CnameRecord{
					Cname: to.Ptr(rec.Target),
				},
			},
		}
		return recordSet, nil
	case libdns.MX:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				MxRecords: []*armdns.MxRecord{{
					Preference: to.Ptr(int32(rec.Preference)),
					Exchange:   to.Ptr(rec.Target),
				}},
			},
		}
		return recordSet, nil
	case libdns.NS:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				NsRecords: []*armdns.NsRecord{{
					Nsdname: to.Ptr(rec.Target),
				}},
			},
		}
		return recordSet, nil
	case libdns.SRV:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				SrvRecords: []*armdns.SrvRecord{{
					Priority: to.Ptr(int32(rec.Priority)),
					Weight:   to.Ptr(int32(rec.Weight)),
					Port:     to.Ptr(int32(rec.Port)),
					Target:   to.Ptr(rec.Target),
				}},
			},
		}
		return recordSet, nil
	case libdns.TXT:
		recordSet := armdns.RecordSet{
			Properties: &armdns.RecordSetProperties{
				TTL: to.Ptr(int64(rec.TTL / time.Second)),
				TxtRecords: []*armdns.TxtRecord{{
					Value: []*string{&rec.Text},
				}},
			},
		}
		return recordSet, nil
	case libdns.RR:
		switch strings.ToUpper(rec.Type) {
		case "PTR":
			recordSet := armdns.RecordSet{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr(int64(rec.TTL / time.Second)),
					PtrRecords: []*armdns.PtrRecord{{
						Ptrdname: to.Ptr(rec.Data),
					}},
				},
			}
			return recordSet, nil
		case "SOA":
			values := strings.Split(rec.Data, " ")
			if len(values) < 7 {
				return armdns.RecordSet{}, fmt.Errorf("invalid SOA record data: %s", rec.Data)
			}

			serialNumber, _ := strconv.ParseInt(values[2], 10, 64)
			refreshTime, _ := strconv.ParseInt(values[3], 10, 64)
			retryTime, _ := strconv.ParseInt(values[4], 10, 64)
			expireTime, _ := strconv.ParseInt(values[5], 10, 64)
			minimumTTL, _ := strconv.ParseInt(values[6], 10, 64)

			recordSet := armdns.RecordSet{
				Properties: &armdns.RecordSetProperties{
					TTL: to.Ptr(int64(rec.TTL / time.Second)),
					SoaRecord: &armdns.SoaRecord{
						Host:         to.Ptr(values[0]),
						Email:        to.Ptr(values[1]),
						SerialNumber: to.Ptr(serialNumber),
						RefreshTime:  to.Ptr(refreshTime),
						RetryTime:    to.Ptr(retryTime),
						ExpireTime:   to.Ptr(expireTime),
						MinimumTTL:   to.Ptr(minimumTTL),
					},
				},
			}
			return recordSet, nil
		default:
			return armdns.RecordSet{}, fmt.Errorf("the type %v cannot be interpreted", rec.Type)
		}
	default:
		return armdns.RecordSet{}, fmt.Errorf("the type %v cannot be interpreted", rec.RR().Type)
	}
}
