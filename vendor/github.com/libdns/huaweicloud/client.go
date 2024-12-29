package huaweicloud

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/libdns/huaweicloud/sdk/core/auth/basic"
	dns "github.com/libdns/huaweicloud/sdk/services/dns/v2"
	"github.com/libdns/huaweicloud/sdk/services/dns/v2/model"
	regions "github.com/libdns/huaweicloud/sdk/services/dns/v2/region"
)

func (p *Provider) getClient() (*dns.DnsClient, error) {
	client := sync.OnceValues(func() (*dns.DnsClient, error) {
		auth, err := basic.NewCredentialsBuilder().
			WithAk(p.AccessKeyId).
			WithSk(p.SecretAccessKey).
			SafeBuild()
		if err != nil {
			return nil, err
		}

		if p.RegionId == "" {
			p.RegionId = "cn-south-1"
		}
		region, err := regions.SafeValueOf(p.RegionId)
		if err != nil {
			return nil, err
		}

		builder, err := dns.DnsClientBuilder().
			WithRegion(region).
			WithCredential(auth).
			SafeBuild()
		if err != nil {
			return nil, err
		}

		return dns.NewDnsClient(builder), nil
	})

	return client()
}

func (p *Provider) getZoneIdByName(name string) (string, error) {
	client, err := p.getClient()
	if err != nil {
		return "", err
	}

	name = strings.TrimSuffix(name, ".")
	searchMode := "equal"
	request := &model.ListPublicZonesRequest{
		Name:       &name,
		SearchMode: &searchMode,
	}

	response, err := client.ListPublicZones(request)
	if err != nil {
		return "", err
	}

	zones := *response.Zones
	if len(zones) == 0 {
		return "", fmt.Errorf("zone %q not found", name)
	}
	if len(zones) != 1 {
		return "", fmt.Errorf("returned more than one zone for %q, expected one, actual %d", name, len(*response.Zones))
	}

	return *zones[0].Id, nil
}

func (p *Provider) getRecordIdByNameAndType(ctx context.Context, zone, recName, recType string) (string, error) {
	records, err := p.GetRecords(ctx, zone)
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if recName == record.Name && recType == record.Type {
			return record.ID, nil
		}
	}

	return "", fmt.Errorf("record %q not found", recName)
}
