package libdnsspaceship

import (
	"encoding/json"
	"strconv"
	"strings"
)

// ResourceRecordBase contains fields common to all Spaceship DNS record payloads
// (this is intentionally minimal; each specific record type augments this with
// fields appropriate for that type).
type ResourceRecordBase struct {
	Type string `json:"type"`
	Name string `json:"name"`
	TTL  int    `json:"ttl,omitempty"`
}

// spaceshipRecordUnion represents the flattened JSON model used by the Spaceship API.
// It contains all possible fields across different DNS record types.

type spaceshipRecordUnion struct {
	ResourceRecordBase

	// type-specific fields (kept flattened for convenience)
	Address    string      `json:"address,omitempty"`
	Cname      string      `json:"cname,omitempty"`
	Value      string      `json:"value,omitempty"`
	Exchange   string      `json:"exchange,omitempty"`
	Preference int         `json:"preference,omitempty"`
	Service    string      `json:"service,omitempty"`
	Protocol   string      `json:"protocol,omitempty"`
	Priority   int         `json:"priority,omitempty"`
	Weight     int         `json:"weight,omitempty"`
	Port       interface{} `json:"port,omitempty"`
	// PortInt is an internal convenience, not serialized
	PortInt         int    `json:"-"`
	Target          string `json:"target,omitempty"`
	Nameserver      string `json:"nameserver,omitempty"`
	Pointer         string `json:"pointer,omitempty"`
	Flag            *int   `json:"flag,omitempty"`
	Tag string `json:"tag,omitempty"`
	// HTTPS/ServiceBinding fields
	SvcPriority     int    `json:"svcPriority,omitempty"`
	SvcTarget       string `json:"svcTarget,omitempty"`
	TargetName      string `json:"targetName,omitempty"`  // Alternative field name used by API
	SvcParams       string `json:"svcParams,omitempty"`
}

// UnmarshalJSON implements custom unmarshalling to handle mixed-type 'port' fields and
// to gracefully decode the API's flattened payloads into the union struct.
func (s *spaceshipRecordUnion) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	// helper to unmarshal if present
	unmarshal := func(key string, v interface{}) {
		if b, ok := raw[key]; ok {
			if err := json.Unmarshal(b, v); err != nil {
				// Log unmarshal errors but don't fail - allows for flexible API evolution
				// In a real application, consider using a proper logger here
			}
		}
	}
	unmarshal("type", &s.Type)
	unmarshal("name", &s.Name)
	unmarshal("ttl", &s.TTL)
	unmarshal("address", &s.Address)
	unmarshal("cname", &s.Cname)
	unmarshal("value", &s.Value)
	unmarshal("exchange", &s.Exchange)
	unmarshal("preference", &s.Preference)
	unmarshal("service", &s.Service)
	// protocol may be present and should be a string
	if b, ok := raw["protocol"]; ok {
		var p string
		if err := json.Unmarshal(b, &p); err == nil {
			s.Protocol = p
		}
	}
	unmarshal("priority", &s.Priority)
	unmarshal("weight", &s.Weight)
	// handle the port value which can be a number or a string (e.g. "_443")
	if b, ok := raw["port"]; ok {
		// try numeric first
		var n int
		if err := json.Unmarshal(b, &n); err == nil {
			s.PortInt = n
			s.Port = n
		} else {
			var ps string
			if err := json.Unmarshal(b, &ps); err == nil {
				s.Port = ps
				if strings.HasPrefix(ps, "_") {
					if v, err := strconv.Atoi(strings.TrimPrefix(ps, "_")); err == nil {
						s.PortInt = v
					}
				} else {
					if v, err := strconv.Atoi(ps); err == nil {
						s.PortInt = v
					}
				}
			}
		}
	}
	unmarshal("target", &s.Target)
	unmarshal("nameserver", &s.Nameserver)
	unmarshal("pointer", &s.Pointer)
	// flag for CAA may be numeric
	if b, ok := raw["flag"]; ok {
		var f int
		if err := json.Unmarshal(b, &f); err == nil {
			s.Flag = &f
		}
	}
	unmarshal("tag", &s.Tag)
	// HTTPS/ServiceBinding fields
	unmarshal("svcPriority", &s.SvcPriority)
	unmarshal("svcTarget", &s.SvcTarget)
	unmarshal("targetName", &s.TargetName)
	unmarshal("svcParams", &s.SvcParams)
	return nil
}
