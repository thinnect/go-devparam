// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "log"
import "time"
import "encoding/hex"
import "encoding/binary"
import "strconv"
import "bytes"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-sfconnection"
import "github.com/thinnect/go-devparam"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 1
const ApplicationVersionPatch = 1

var ApplicationBuildDate string
var ApplicationBuildDistro string

func parseValue(opts Options) ([]byte, error) {
	var t interface{}
	if opts.Uint8 {
		if v, err := strconv.ParseUint(opts.Value, 10, 8); err == nil {
			t = uint8(v)
		} else {
			return nil, err
		}
	} else if opts.Uint16 {
		if v, err := strconv.ParseUint(opts.Value, 10, 16); err == nil {
			t = uint16(v)
		} else {
			return nil, err
		}
	} else if opts.Uint32 {
		if v, err := strconv.ParseUint(opts.Value, 10, 32); err == nil {
			t = uint32(v)
		} else {
			return nil, err
		}
	} else if opts.Uint64 {
		if v, err := strconv.ParseUint(opts.Value, 10, 64); err == nil {
			t = uint64(v)
		} else {
			return nil, err
		}
	} else if opts.Int8 {
		if v, err := strconv.ParseInt(opts.Value, 10, 8); err == nil {
			t = int8(v)
		} else {
			return nil, err
		}
	} else if opts.Int16 {
		if v, err := strconv.ParseInt(opts.Value, 10, 16); err == nil {
			t = int16(v)
		} else {
			return nil, err
		}
	} else if opts.Int32 {
		if v, err := strconv.ParseInt(opts.Value, 10, 32); err == nil {
			t = int32(v)
		} else {
			return nil, err
		}
	} else if opts.Int64 {
		if v, err := strconv.ParseInt(opts.Value, 10, 64); err == nil {
			t = int64(v)
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
			return buf.Bytes(), nil
		}
	}

	if opts.String {
		return []byte(opts.Value), nil
	}

	return hex.DecodeString(opts.Value)
}

type Options struct {
	Positional struct {
		ConnectionString string `description:"Connectionstring sf@HOST:PORT"`
	} `positional-args:"yes"`

	Group       sfconnection.AMGroup `short:"g" long:"group" default:"22" description:"Packet AM Group (hex)"`
	Address     sfconnection.AMAddr  `short:"a" long:"address" default:"5678" description:"Source AM address (hex)"`
	Destination sfconnection.AMAddr  `short:"d" long:"destination" default:"0" description:"Destination AM address (hex)"`

	Timeout int `long:"timeout" default:"1" description:"Get/set action timeout (seconds)"`
	Retries int `long:"retries" default:"3" description:"Get/set action retries"`

	Parameter []string `short:"p" long:"parameter" description:"List of parameter names"`
	Value     string   `short:"v" long:"value"     description:"Value to set (single parameter only)"`

	String bool `long:"string" description:"Value is string"`
	Uint8  bool `long:"uint8"  description:"Value is uint8"`
	Uint16 bool `long:"uint16" description:"Value is uint16"`
	Uint32 bool `long:"uint32" description:"Value is uint32"`
	Uint64 bool `long:"uint64" description:"Value is uint64"`
	Int8   bool `long:"int8"   description:"Value is int8"`
	Int16  bool `long:"int16"  description:"Value is int16"`
	Int32  bool `long:"int32"  description:"Value is int32"`
	Int64  bool `long:"int64"  description:"Value is int64"`

	Debug       []bool `short:"D" long:"debug"   description:"Debug mode, print raw packets"`
	Quiet       []bool `short:"Q" long:"quiet"   description:"Quiet mode, print only values"`
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

	host, port, err := sfconnection.ParseSfConnectionString(opts.Positional.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	sfc := sfconnection.NewSfConnection()

	var dpm *deviceparameters.DeviceParameterManager = nil
	if opts.Destination == 0 {
		dpm = deviceparameters.NewDeviceParameterManager(sfc)
	} else {
		dpm = deviceparameters.NewDeviceParameterActiveMessageManager(sfc, opts.Group, opts.Address, opts.Destination)
	}
	dpm.SetTimeout(time.Duration(opts.Timeout) * time.Second)
	dpm.SetRetries(opts.Retries)

	logger := logsetup(len(opts.Debug))
	if len(opts.Debug) > 0 {
		sfc.SetLoggers(logger)
	}
	dpm.SetLoggers(logger)

	// Connect to the host
	err = sfc.Connect(host, port)
	if err != nil {
		logger.Error.Printf("Unable to connect to %s:%d\n", host, port)
		os.Exit(1)
	}
	if len(opts.Quiet) == 0 {
		logger.Info.Printf("Connected to %s:%d\n", host, port)
	}

	if len(opts.Parameter) > 0 {
		if len(opts.Value) == 0 || len(opts.Parameter) > 1 {
			for _, parameter := range opts.Parameter {
				if len(opts.Quiet) == 0 {
					logger.Info.Printf("Get %s\n", parameter)
				}
				val, err := dpm.GetValue(parameter)
				if err == nil {
					logger.Info.Printf("%s = %s\n", val.Name, val)
				} else {
					logger.Info.Printf("Failed: %s\n", err)
				}
			}
		} else { // Set only if value and only a single parameter
			value, err := parseValue(opts)
			if err == nil {
				logger.Info.Printf("Set %s to 0x%X\n", opts.Parameter[0], value)
				if val, err := dpm.SetValue(opts.Parameter[0], value); err == nil {
					logger.Info.Printf("%s = %s\n", val.Name, val)
				} else {
					logger.Info.Printf("Failed: %s\n", err)
				}
			} else {
				logger.Error.Printf("%s", err)
			}
		}
	} else {
		logger.Info.Printf("Get parameter list:\n")
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
		} else {
			logger.Info.Printf("Failed: %s\n", err)
		}
	}

	dpm.Close()
	sfc.Disconnect()
	time.Sleep(100 * time.Millisecond)
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
