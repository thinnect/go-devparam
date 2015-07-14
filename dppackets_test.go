// Author  Raido Pahtma
// License MIT

package deviceparameters

import "testing"
import "fmt"
import "encoding/hex"
import "github.com/proactivity-lab/go-sfconnection"

func TestSerializer(t *testing.T) {
	var p DpParameter
	p.Header = DP_PARAMETER
	p.Seqnum = 0
	p.Id = "test"
	p.Value = []byte{1, 2, 3}
	fmt.Printf("np %v\n", p)
	b := sfconnection.SerializePacket(&p)
	fmt.Printf("sp %v\n", p)
	fmt.Printf("rb %X\n", b)
}

func TestDeserializer(t *testing.T) {
	var dp DpParameter

	raw, _ := hex.DecodeString("1000040374657374010203")
	if err := sfconnection.DeserializePacket(&dp, raw); err != nil {
		t.Error("error %s", err)
	}

	fmt.Printf("dp %v\n", dp)
}
