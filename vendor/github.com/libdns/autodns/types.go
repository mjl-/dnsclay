package autodns

import (
	"fmt"
	"strings"
)

type AutoDNSObject struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type AutoDNSUser struct {
	Context int32  `json:"context"`
	User    string `json:"user"`
}

type RequestZone struct {
	Domain string `json:"domain"`
}

type AutoDNSMessage struct {
	Text    string          `json:"text"`
	Objects []AutoDNSObject `json:"objects"`
	Code    string          `json:"code"`
	Status  string          `json:"status"`
}

type AutoDNSResponse struct {
	STID string `json:"stid"`

	Status struct {
		Type string  `json:"type"`
		Code *string `json:"resultCode,omitempty"`
		Text *string `json:"text,omitempty"`
	} `json:"status"`

	Object *AutoDNSObject `json:"object,omitempty"`

	// potential error messages
	Messages []*AutoDNSMessage `json:"messages,omitempty"`
}

type ResponseSearch struct {
	AutoDNSResponse
	Data []ResponseSearchItem `json:"data"`
}

type ResponseZone struct {
	AutoDNSResponse
	Data []ZoneItem `json:"data"`
}

// {
// 	"created": "2023-10-18T13:56:47.000+0200",
// 	"updated": "2024-10-25T13:16:43.000+0200",
// 	"origin": "something.example.org",
// 	"nameServerGroup": "ns14.net",
// 	"owner": {
// 	  "context": 4,
// 	  "user": "user"
// 	},
// 	"updater": {
// 	  "context": 4,
// 	  "user": "user"
// 	},
// 	"domainsafe": false,
// 	"wwwInclude": false,
// 	"virtualNameServer": "a.ns14.net"
// }

type ResponseSearchItem struct {
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	Origin      string      `json:"origin"`
	NSGroup     string      `json:"nameServerGroup"`
	Owner       AutoDNSUser `json:"owner"`
	Updater     AutoDNSUser `json:"updater"`
	DomainSafe  bool        `json:"domainsafe"`
	WWWWInclude bool        `json:"wwwInclude"`
	Nameserver  string      `json:"virtualNameserver"`
}

//	{
//		"created": "2023-10-18T13:56:47.000+0200",
//		"updated": "2024-10-25T13:16:43.000+0200",
//		"origin": "something.example.org",
//		"soa": {
//		  "refresh": 43200,
//		  "retry": 7200,
//		  "expire": 1209600,
//		  "ttl": 86400,
//		  "email": "do-not-reply@something.example.org"
//		},
//		"nameServerGroup": "ns14.net",
//		"owner": {
//		  "context": 4,
//		  "user": "user"
//		},
//		"updater": {
//		  "context": 4,
//		  "user": "user"
//		},
//		"domainsafe": false,
//		"purgeType": "DISABLED",
//		"nameServers": [
//		  {
//			"name": "a.ns14.net"
//		  },
//		  {
//			"name": "b.ns14.net"
//		  },
//		  {
//			"name": "c.ns14.net"
//		  },
//		  {
//			"name": "d.ns14.net"
//		  }
//		],
//		"main": {
//		  "address": "127.0.0.1"
//		},
//		"wwwInclude": false,
//		"virtualNameServer": "a.ns14.net",
//		"action": "COMPLETE",
//		"resourceRecords": [
//		  {
//			"name": "*",
//			"type": "A",
//			"value": "127.0.0.1"
//		  }
//		],
//		"roid": 9149383
//	  }
//
// ]

type ZoneRecord struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ZoneItem struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
	Origin  string `json:"origin"`

	SOA struct {
		Refresh int    `json:"refresh"`
		Retry   int    `json:"retry"`
		Expire  int    `json:"expire"`
		TTL     int    `json:"ttl"`
		Email   string `json:"email"`
	} `json:"soa"`

	NSGroup    string      `json:"nameServerGroup"`
	Owner      AutoDNSUser `json:"owner"`
	Updater    AutoDNSUser `json:"updater"`
	DomainSafe bool        `json:"domainsafe"`
	PurgeType  string      `json:"purgeType"`

	Nameservers []struct {
		Name string `json:"name"`
	} `json:"nameservers"`

	Main struct {
		Address string `json:"address"`
	} `json:"main"`

	WWWWInclude bool   `json:"wwwInclude"`
	Nameserver  string `json:"virtualNameserver"`
	Action      string `json:"action"`

	Records []ZoneRecord `json:"resourceRecords"`

	ROID int `json:"roid"`
}

type AutoDNSError struct {
	messages []*AutoDNSMessage
}

func (m *AutoDNSError) Error() string {
	if m.messages == nil {
		return "unknown error"
	}

	var errs []string
	for _, m := range m.messages {
		objects := []string{}
		for _, o := range m.Objects {
			objects = append(objects, "%s (type: %s)", o.Value, o.Type)
		}

		errs = append(errs, fmt.Sprintf("%s, code: %s, objects: %s",
			m.Text, m.Code, strings.Join(objects, ", "),
		))
	}

	return strings.Join(errs, "; ")
}

func (m *AutoDNSError) Messages() []*AutoDNSMessage {
	return m.messages
}

func NewError(resp AutoDNSResponse) *AutoDNSError {
	return &AutoDNSError{
		messages: resp.Messages,
	}
}
