// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "log"
import "time"
import "errors"
import "encoding/hex"
import "encoding/binary"
import "strconv"
import "bytes"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-moteconnection"
import "github.com/thinnect/go-devparam"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 2
const ApplicationVersionPatch = 1

var ApplicationBuildDate string
var ApplicationBuildDistro string

func parseValue(opts Options) ([]byte, error) {
	var value []byte
	var err error
	var t interface{}
	c := 0

	if len(opts.Uint8) > 0 {
		if v, err := strconv.ParseUint(opts.Uint8, 10, 8); err == nil {
			t = uint8(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Uint16) > 0 {
		if v, err := strconv.ParseUint(opts.Uint16, 10, 16); err == nil {
			t = uint16(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Uint32) > 0 {
		if v, err := strconv.ParseUint(opts.Uint32, 10, 32); err == nil {
			t = uint32(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Uint64) > 0 {
		if v, err := strconv.ParseUint(opts.Uint64, 10, 64); err == nil {
			t = uint64(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Int8) > 0 {
		if v, err := strconv.ParseInt(opts.Int8, 10, 8); err == nil {
			t = int8(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Int16) > 0 {
		if v, err := strconv.ParseInt(opts.Int16, 10, 16); err == nil {
			t = int16(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Int32) > 0 {
		if v, err := strconv.ParseInt(opts.Int32, 10, 32); err == nil {
			t = int32(v)
			c++
		} else {
			return nil, err
		}
	}
	if len(opts.Int64) > 0 {
		if v, err := strconv.ParseInt(opts.Int64, 10, 64); err == nil {
			t = int64(v)
			c++
		} else {
			return nil, err
		}
	}

	if t != nil {
		switch t := t.(type) {
		case uint8, uint16, uint32, uint64, int8, int16, int32, int64:
			buf := new(bytes.Buffer)
			if err := binary.Write(buf, binary.BigEndian, t); err != nil {
				return nil, err
			}
			value = buf.Bytes()
		}
	}

	if len(opts.Value) > 0 {
		value, err = hex.DecodeString(opts.Value)
		c++
	}

	if len(opts.String) > 0 {
		value = []byte(opts.String)
		c++
	}

	if c > 1 {
		return nil, errors.New("Multiple values specified for parameter!")
	}

	return value, err
}

type Options struct {
	Positional struct {
		ConnectionString string `description:"Connectionstring sf@HOST:PORT or serial@PORT:BAUD"`
	} `positional-args:"yes"`

	Group       moteconnection.AMGroup `short:"g" long:"group" default:"22" description:"Packet AM Group (hex)"`
	Address     moteconnection.AMAddr  `short:"a" long:"address" default:"5678" description:"Source AM address (hex)"`
	Destination moteconnection.AMAddr  `short:"d" long:"destination" default:"0" description:"Destination AM address (hex)"`

	Timeout int `long:"timeout" default:"1" description:"Get/set action timeout (seconds)"`
	Retries int `long:"retries" default:"3" description:"Get/set action retries"`

	Parameter []string `short:"p" long:"parameter" description:"List of parameter names"`

	Value  string `short:"v" long:"value"     description:"Set value, presented as a raw hex buffer"`
	String string `long:"str" description:"Set value, type is string"`
	Uint8  string `long:"u8"  description:"Set value, type is uint8"`
	Uint16 string `long:"u16" description:"Set value, type is uint16"`
	Uint32 string `long:"u32" description:"Set value, type is uint32"`
	Uint64 string `long:"u64" description:"Set value, type is uint64"`
	Int8   string `long:"i8"  description:"Set value, type is int8"`
	Int16  string `long:"i16" description:"Set value, type is int16"`
	Int32  string `long:"i32" description:"Set value, type is int32"`
	Int64  string `long:"i64" description:"Set value, type is int64"`

	Quiet       []bool `short:"Q" long:"quiet"   description:"Quiet mode, print only values"`
	Debug       []bool `short:"D" long:"debug"   description:"Debug mode, print raw packets"`
	ShowVersion func() `short:"V" long:"version" description:"Show application version"`
}

func main() {

	var opts Options
	opts.ShowVersion = func() {
		if ApplicationBuildDate == "" {
			ApplicationBuildDate = "YYYY-mm-dd_HH:MM:SS"
		}
		if ApplicationBuildDistro == "" {
			ApplicationBuildDistro = "unknown"
		}
		fmt.Printf("deviceparameter %d.%d.%d (%s %s)\n", ApplicationVersionMajor, ApplicationVersionMinor, ApplicationVersionPatch, ApplicationBuildDate, ApplicationBuildDistro)
		os.Exit(0)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf("Argument parser error: %s\n", err)
		os.Exit(1)
	}

	conn, cs, err := moteconnection.CreateConnection(opts.Positional.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	var dpm *deviceparameters.DeviceParameterManager = nil
	if opts.Destination == 0 {
		dpm = deviceparameters.NewDeviceParameterManager(conn)
	} else {
		dpm = deviceparameters.NewDeviceParameterActiveMessageManager(conn, opts.Group, opts.Address, opts.Destination)
	}
	dpm.SetTimeout(time.Duration(opts.Timeout) * time.Second)
	dpm.SetRetries(opts.Retries)

	logger := logsetup(len(opts.Debug))
	if len(opts.Debug) > 0 {
		conn.SetLoggers(logger)
	}
	dpm.SetLoggers(logger)

	err = conn.Connect()
	if err != nil {
		logger.Error.Printf("Unable to connect with %s: %s\n", cs, err)
		os.Exit(1)
	}
	if len(opts.Quiet) == 0 {
		logger.Info.Printf("Connected with %s\n", cs)
	}

	success := false

	if len(opts.Parameter) > 0 {
		value, err := parseValue(opts)
		if err != nil {
			logger.Error.Printf("%s", err)
		} else if value != nil && len(opts.Parameter) > 1 {
			logger.Error.Printf("Value and multiple parameters provided\n")
		} else if value == nil {
			for _, parameter := range opts.Parameter {
				if len(opts.Quiet) == 0 {
					logger.Info.Printf("Get %s\n", parameter)
				}
				val, err := dpm.GetValue(parameter)
				if err == nil {
					logger.Info.Printf("%s = %s\n", val.Name, val)
					success = true
				} else {
					logger.Info.Printf("Failed: %s\n", err)
				}
			}
		} else { // Set only if value and only a single parameter
			logger.Info.Printf("Set %s to 0x%X\n", opts.Parameter[0], value)
			if val, err := dpm.SetValue(opts.Parameter[0], value); err == nil {
				logger.Info.Printf("%s = %s\n", val.Name, val)
				success = true
			} else {
				logger.Info.Printf("Failed: %s\n", err)
			}
		}
	} else {
		if len(opts.Quiet) == 0 {
			logger.Info.Printf("Get parameter list:\n")
		}
		pchan, err := dpm.GetList()
		if err == nil {
			param := <-pchan
			for ; param != nil; param = <-pchan {
				if param.Error == nil {
					logger.Info.Printf("%2d: %s %s\n", param.Seqnum, param.Name, param)
				} else {
					logger.Info.Printf("%2d: %s\n", param.Seqnum, param.Error)
				}
			}
			success = true
		} else {
			logger.Info.Printf("Failed: %s\n", err)
		}
	}

	dpm.Close()
	conn.Disconnect()
	time.Sleep(100 * time.Millisecond)

	if success {
		if len(opts.Quiet) == 0 {
			logger.Info.Printf("Done")
		}
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func logsetup(debuglevel int) *loggers.DIWEloggers {
	logger := loggers.New()
	logformat := log.Ldate | log.Ltime | log.Lmicroseconds

	if debuglevel > 1 {
		logformat = logformat | log.Lshortfile
	}

	if debuglevel > 0 {
		logger.SetDebugLogger(log.New(os.Stdout, "DEBUG: ", logformat))
		logger.SetInfoLogger(log.New(os.Stdout, "INFO:  ", logformat))
	} else {
		logger.SetInfoLogger(log.New(os.Stdout, "", logformat))
	}
	logger.SetWarningLogger(log.New(os.Stdout, "WARN:  ", logformat))
	logger.SetErrorLogger(log.New(os.Stdout, "ERROR: ", logformat))
	return logger
}
