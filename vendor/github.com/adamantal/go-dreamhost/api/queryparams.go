package api

type baseParams struct {
	APIKey  string `url:"key" json:"key"`
	Command string `url:"cmd" json:"command"`
	Format  string `url:"format,omitempty" json:"format"`
}

type dnsRecordParams struct {
	baseParams     `url:",inline" json:",inline"`
	DNSRecordInput `url:",inline" json:",inline"`
}
