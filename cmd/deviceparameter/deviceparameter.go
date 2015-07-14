// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "log"
import "time"
import "encoding/hex"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-sfconnection"
import "github.com/thinnect/go-devparam"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 1
const ApplicationVersionPatch = 0

var ApplicationBuildDate string
var ApplicationBuildDistro string

type HexString []byte

func (self *HexString) UnmarshalFlag(value string) error {
	data, err := hex.DecodeString(value)
	*self = data
	return err
}

func (self HexString) MarshalFlag() (string, error) {
	return hex.EncodeToString(self), nil
}

func main() {

	var opts struct {
		Positional struct {
			ConnectionString string `description:"Connectionstring sf@HOST:PORT"`
		} `positional-args:"yes"`

		Group sfconnection.AMGroup `short:"g" long:"group" default:"22" description:"Packet AM Group (hex)"`

		Parameter string    `short:"p" long:"parameter" description:"Name of the parameter"`
		Value     HexString `short:"v" long:"value"     description:"Value to set"`

		Debug       []bool `short:"D" long:"debug"   description:"Debug mode, print raw packets"`
		ShowVersion func() `short:"V" long:"version" description:"Show application version"`
	}

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
	dpm := deviceparameters.NewDeviceParameterManager(sfc)

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
	logger.Info.Printf("Connected to %s:%d\n", host, port)

	if len(opts.Parameter) > 0 {
		if len(opts.Value) > 0 {
			logger.Info.Printf("Set %s to %X\n", opts.Parameter, opts.Value)
			err := dpm.SetValue(opts.Parameter, opts.Value)
			if err == nil {
				logger.Info.Printf("Done\n")
			} else {
				logger.Info.Printf("Failed: %s\n", err)
			}
		} else {
			logger.Info.Printf("Get %s\n", opts.Parameter)
			val, err := dpm.GetValue(opts.Parameter)
			if err == nil {
				logger.Info.Printf("%s = %X\n", opts.Parameter, val)
			} else {
				logger.Info.Printf("Failed: %s\n", err)
			}
		}
	} else {
		logger.Info.Printf("Get parameter list:\n")
		pchan, err := dpm.GetList()
		if err == nil {
			param := <-pchan
			for ; param != nil; param = <-pchan {
				if param.Error == nil {
					logger.Info.Printf("%d: %s %s\n", param.Seqnum, param.Name, param)
				} else {
					logger.Info.Printf("%d: %s\n", param.Seqnum, param.Error)
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
