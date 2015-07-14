// Author  Raido Pahtma
// License MIT

package deviceparameters

const DP_HEARTBEAT = 0x00
const DP_PARAMETER = 0x10
const DP_GET_PARAMETER_WITH_ID = 0x21
const DP_GET_PARAMETER_WITH_SEQNUM = 0x22

const DP_SET_PARAMETER_WITH_ID = 0x31
const DP_SET_PARAMETER_WITH_SEQNUM = 0x32

const DP_ERROR_PARAMETER_ID = 0xF0
const DP_ERROR_PARAMETER_SEQNUM = 0xF1

type DpHeartbeat struct {
	Header uint8
	Eui64  uint64
	Uptime uint32
}

type DpParameter struct {
	Header      uint8
	Type        uint8
	Seqnum      uint8
	IdLength    uint8 `sfpacket:"len(Id)"`
	ValueLength uint8 `sfpacket:"len(Value)"`
	Id          string
	Value       []byte
}

type DpGetParameterSeqnum struct {
	Header uint8
	Seqnum uint8
}

type DpGetParameterId struct {
	Header   uint8
	IdLength uint8 `sfpacket:"len(Id)"`
	Id       string
}

type DpSetParameterSeqnum struct {
	Header      uint8
	Seqnum      uint8
	ValueLength uint8 `sfpacket:"len(Value)"`
	Value       []byte
}

type DpSetParameterId struct {
	Header      uint8
	IdLength    uint8 `sfpacket:"len(Id)"`
	ValueLength uint8 `sfpacket:"len(Value)"`
	Id          string
	Value       []byte
}

type DpErrorParameterSeqnum struct {
	Header uint8
	Exists bool
	Err    uint8
	Seqnum uint8
}

type DpErrorParameterId struct {
	Header   uint8
	Exists   bool
	Err      uint8
	IdLength uint8 `sfpacket:"len(Id)"`
	Id       string
}
