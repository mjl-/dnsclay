package model

import (
	"strings"

	"github.com/libdns/huaweicloud/sdk/core/utils"
)

// DeleteRecordSetRequest Request Object
type DeleteRecordSetRequest struct {

	// 当前recordset所属的zoneID。
	ZoneId string `json:"zone_id"`

	// 当前recordset所属的ID信息。
	RecordsetId string `json:"recordset_id"`
}

func (o DeleteRecordSetRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteRecordSetRequest struct{}"
	}

	return strings.Join([]string{"DeleteRecordSetRequest", string(data)}, " ")
}
