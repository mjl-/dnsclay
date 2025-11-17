package huaweicloud

import (
	"time"

	"github.com/libdns/libdns"
)

type ListZonesResponse struct {
	Zones []Zone `json:"zones,omitempty"`
}

type ListRecordsResponse struct {
	RecordSets []RecordSet `json:"recordsets,omitempty"`
}

type Zone struct {
	// zone的ID，uuid形式的一个资源标识。
	Id string `json:"id,omitempty"`
	// zone名称。
	Name string `json:"name,omitempty"`
}

type RecordSet struct {
	// Record Set的ID
	Id string `json:"id,omitempty"`
	// Record Set的名称（FQDN形式）
	Name string `json:"name,omitempty"`
	// 记录类型。
	Type string `json:"type,omitempty"`
	// 解析记录在本地DNS服务器的缓存时间，缓存时间越长更新生效越慢，以秒为单位。
	Ttl int32 `json:"ttl,omitempty"`
	// 域名解析后的值。
	Records []string `json:"records,omitempty"`
}

func (r RecordSet) libdnsRecord(zone string) ([]libdns.Record, error) {
	var records []libdns.Record
	for _, record := range r.Records {
		rr, err := libdns.RR{
			Name: libdns.RelativeName(r.Name, zone),
			TTL:  time.Duration(r.Ttl) * time.Second,
			Type: r.Type,
			Data: record,
		}.Parse()
		if err != nil {
			return nil, err
		}
		records = append(records, rr)
	}
	return records, nil
}

func hwRecord(zone string, r libdns.Record) (RecordSet, error) {
	rr := r.RR()
	if rr.TTL <= 0 {
		rr.TTL = 1 * time.Second // 华为云支持最小 1 秒
	}
	// Fix TXT records missing starting quotation mark
	// https://github.com/libdns/huaweicloud/issues/1
	if rr.Type == "TXT" {
		if rr.Data[0] != '"' {
			rr.Data = `"` + rr.Data
		}
		if rr.Data[len(rr.Data)-1] != '"' {
			rr.Data = rr.Data + `"`
		}
	}
	return RecordSet{
		Name:    libdns.AbsoluteName(rr.Name, zone),
		Type:    rr.Type,
		Ttl:     int32(rr.TTL.Seconds()),
		Records: []string{rr.Data},
	}, nil
}
