// Author  Raido Pahtma
// License MIT

package deviceparameters

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/proactivity-lab/go-moteconnection"
)

func TestDpm(t *testing.T) {
	sfc := moteconnection.NewSfConnection("localhost", 9002)
	dp := NewDeviceParameterManager(sfc)

	logformat := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	debuglogger := log.New(os.Stdout, "DEBUG: ", logformat)
	infologger := log.New(os.Stdout, "INFO:  ", logformat)
	warninglogger := log.New(os.Stdout, "WARN:  ", logformat)
	errorlogger := log.New(os.Stdout, "ERROR: ", logformat)

	sfc.SetDebugLogger(debuglogger)
	sfc.SetInfoLogger(infologger)
	sfc.SetWarningLogger(warninglogger)
	sfc.SetErrorLogger(errorlogger)

	dp.SetDebugLogger(debuglogger)
	dp.SetInfoLogger(infologger)
	dp.SetWarningLogger(warninglogger)
	dp.SetErrorLogger(errorlogger)

	sfc.Autoconnect(30 * time.Second)

	time.Sleep(time.Second)

	v1, err := dp.GetValue("radio_channel")
	fmt.Printf("%v %v\n", v1, err)

	v2, err := dp.GetValue("ident_timestamp")
	fmt.Printf("%v %v\n", v2, err)

	v3, err := dp.GetValue("dummy")
	fmt.Printf("%v %v\n", v3, err)

	dp.SetTimeout(0)
	v4, err := dp.GetValue("dummy")
	fmt.Printf("%v %v\n", v4, err)

	dp.SetTimeout(time.Second)
	_, err = dp.SetValue("radio_channel", []byte{11})
	fmt.Printf("s %v\n", err)

	dp.Close()
	sfc.Disconnect()
}
