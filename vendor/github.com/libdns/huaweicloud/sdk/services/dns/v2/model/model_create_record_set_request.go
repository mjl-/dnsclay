package model

import (
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/utils"
)

// CreateRecordSetRequest Request Object
type CreateRecordSetRequest struct {

	// 所属zone的ID。
	ZoneId string `json:"zone_id"`

	Body *CreateRecordSetRequestBody `json:"body,omitempty"`
}

func (o CreateRecordSetRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateRecordSetRequest struct{}"
	}

	return strings.Join([]string{"CreateRecordSetRequest", string(data)}, " ")
}
