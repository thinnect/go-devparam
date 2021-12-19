// Author  Raido Pahtma
// License MIT

package deviceparameters

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/proactivity-lab/go-moteconnection"
)

func TestSerializer(t *testing.T) {
	var p DpParameter
	p.Header = DP_PARAMETER
	p.Seqnum = 0
	p.Id = "test"
	p.Value = []byte{1, 2, 3}
	fmt.Printf("np %v\n", p)
	b := moteconnection.SerializePacket(&p)
	fmt.Printf("sp %v\n", p)
	fmt.Printf("rb %X\n", b)
}

func TestDeserializer(t *testing.T) {
	var dp DpParameter

	s := strings.Replace("10 00 04 04 03 74657374 010203", " ", "", -1)
	raw, _ := hex.DecodeString(s)
	if err := moteconnection.DeserializePacket(&dp, raw); err != nil {
		t.Errorf("error %s", err)
	}

	fmt.Printf("dp %v\n", dp)
}
