package model

import (
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/utils"
)

// ListRecordSetsByZoneResponse Response Object
type ListRecordSetsByZoneResponse struct {
	Links *PageLink `json:"links,omitempty"`

	// recordset列表对象。
	Recordsets *[]ListRecordSets `json:"recordsets,omitempty"`

	Metadata       *Metadata `json:"metadata,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListRecordSetsByZoneResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListRecordSetsByZoneResponse struct{}"
	}

	return strings.Join([]string{"ListRecordSetsByZoneResponse", string(data)}, " ")
}
