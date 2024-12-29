package huaweicloud

func prepareRecordValue(rType string, value string) []string {
	switch rType {
	case "TXT":
		// Fix TXT records missing starting quotation mark
		// https://github.com/libdns/huaweicloud/issues/1
		if value[0] != '"' {
			value = `"` + value
		}
		if value[len(value)-1] != '"' {
			value = value + `"`
		}
	}

	return []string{value}
}
