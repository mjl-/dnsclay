package zones

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mittwald/go-powerdns/pdnshttp"
)

func (c *client) RemoveRecordSetFromZone(ctx context.Context, serverID string, zoneID string, name string, recordType string) error {
	set := ResourceRecordSet{
		Name:       name,
		Type:       recordType,
		ChangeType: ChangeTypeDelete,
	}

	return c.RemoveRecordSetsFromZone(ctx, serverID, zoneID, []ResourceRecordSet{set})
}

func (c *client) RemoveRecordSetsFromZone(ctx context.Context, serverID string, zoneID string, sets []ResourceRecordSet) error {
	path := fmt.Sprintf("/servers/%s/zones/%s", url.PathEscape(serverID), url.PathEscape(zoneID))

	for idx := range sets {
		sets[idx].ChangeType = ChangeTypeDelete
	}
	patch := Zone{
		ResourceRecordSets: sets,
	}

	return c.httpClient.Patch(ctx, path, nil, pdnshttp.WithJSONRequestBody(&patch))
}
