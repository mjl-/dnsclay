package googleclouddns

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/dns/v1"
	"google.golang.org/api/option"
)

const (
	// zoneMapTTL timeout for the Google Cloud DNS zone map
	zoneMapTTL = time.Minute * 5
)

// newService initializes the Google client for the provider using the specified JSON file for credentials if set.
func (p *Provider) newService(ctx context.Context) error {
	var err error
	if p.service == nil {
		scopeOption := option.WithScopes(dns.NdevClouddnsReadwriteScope)
		if p.ServiceAccountJSON != "" {
			p.service, err = dns.NewService(ctx, scopeOption, option.WithCredentialsFile(p.ServiceAccountJSON))
		} else {
			p.service, err = dns.NewService(ctx, scopeOption)
		}
	}
	return err
}

// getCloudDNSZone will return the Google Cloud DNS zone name for the specified zone. The data is cached
// for five minutes to avoid repeated calls to the GCP API servers.
func (p *Provider) getCloudDNSZone(zone string) (string, error) {
	if p.zoneMap == nil || time.Since(p.zoneMapLastUpdated) > zoneMapTTL {
		p.zoneMap = make(map[string]string)
		zonesLister := p.service.ManagedZones.List(p.Project)
		err := zonesLister.Pages(context.Background(), func(response *dns.ManagedZonesListResponse) error {
			for _, zone := range response.ManagedZones {
				if zone.Visibility == "public" {
					p.zoneMap[zone.DnsName] = zone.Name
				}
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		p.zoneMapLastUpdated = time.Now()
	}
	if zoneName, ok := p.zoneMap[zone]; ok {
		return zoneName, nil
	}
	return "", fmt.Errorf("unable to find Google managaged zone for domain %s", zone)
}
