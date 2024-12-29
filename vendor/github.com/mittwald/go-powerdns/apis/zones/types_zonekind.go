package zones

import "fmt"

type ZoneKind int

const (
	_                       = iota
	ZoneKindNative ZoneKind = iota
	ZoneKindMaster
	ZoneKindSlave
	ZoneKindProducer
	ZoneKindConsumer
)

func (k ZoneKind) MarshalJSON() ([]byte, error) {
	switch k {
	case ZoneKindNative:
		return []byte(`"Native"`), nil
	case ZoneKindMaster:
		return []byte(`"Master"`), nil
	case ZoneKindSlave:
		return []byte(`"Slave"`), nil
	case ZoneKindProducer:
		return []byte(`"Producer"`), nil
	case ZoneKindConsumer:
		return []byte(`"Consumer"`), nil
	default:
		return nil, fmt.Errorf("unsupported zone kind: %d", k)
	}
}

func (k *ZoneKind) UnmarshalJSON(input []byte) error {
	switch string(input) {
	case `"Native"`:
		*k = ZoneKindNative
	case `"Master"`:
		*k = ZoneKindMaster
	case `"Slave"`:
		*k = ZoneKindSlave
	case `"Producer"`:
		*k = ZoneKindProducer
	case `"Consumer"`:
		*k = ZoneKindConsumer
	default:
		return fmt.Errorf("unsupported zone kind: %s", string(input))
	}

	return nil
}
