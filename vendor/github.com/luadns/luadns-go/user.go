package luadns

type User struct {
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	RepoURI     string   `json:"repo_uri"`
	APIEnabled  bool     `json:"api_enabled"`
	TFA         bool     `json:"tfa"`
	DeployKey   string   `json:"deploy_key"`
	TTL         uint32   `json:"ttl"`
	Package     string   `json:"package"`
	NameServers []string `json:"name_servers"`
}
