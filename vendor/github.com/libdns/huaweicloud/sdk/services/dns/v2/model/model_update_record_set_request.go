package model

import (
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/utils"
)

// UpdateRecordSetRequest Request Object
type UpdateRecordSetRequest struct {

	// 所属zone的ID。
	ZoneId string `json:"zone_id"`

	// 待修改的recordset的ID信息。
	RecordsetId string `json:"recordset_id"`

	Body *UpdateRecordSetReq `json:"body,omitempty"`
}

func (o UpdateRecordSetRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "UpdateRecordSetRequest struct{}"
	}

	return strings.Join([]string{"UpdateRecordSetRequest", string(data)}, " ")
}
