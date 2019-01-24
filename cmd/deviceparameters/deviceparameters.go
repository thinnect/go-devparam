// Author  Raido Pahtma
// License MIT

package main

import "fmt"
import "os"
import "log"
import "time"
import "os/signal"

//import "errors"
//import "encoding/hex"
//import "encoding/binary"
//import "strconv"
//import "bytes"

import "github.com/jessevdk/go-flags"
import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-moteconnection"
import "github.com/thinnect/go-devparam/director"

const ApplicationVersionMajor = 0
const ApplicationVersionMinor = 3
const ApplicationVersionPatch = 0

var ApplicationBuildDate string
var ApplicationBuildDistro string

type Options struct {
	Positional struct {
		File string `description:"Parameter configuration work file." required:"true"`
	} `positional-args:"yes"`

	ConnectionString string `long:"conn" default:"sf@localhost:9002" description:"Connectionstring sf@HOST:PORT or serial@PORT:BAUD"`

	Group   moteconnection.AMGroup `short:"g" long:"group" default:"22" description:"Packet AM Group (hex)"`
	Address moteconnection.AMAddr  `short:"a" long:"address" default:"5678" description:"Source AM address (hex)"`

	Template string `short:"t" long:"template" default:"" description:"Template for activities."`
	List     string `short:"l" long:"list" default:"" description:"List of nodes to apply the template for."`

	Timeout int   `long:"timeout" default:"10" description:"Get/set action timeout (seconds)"`
	Retries uint8 `long:"retries" default:"3" description:"Get/set action retries"`

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
		fmt.Printf("deviceparameters %d.%d.%d (%s %s)\n", ApplicationVersionMajor, ApplicationVersionMinor, ApplicationVersionPatch, ApplicationBuildDate, ApplicationBuildDistro)
		os.Exit(0)
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Printf("Argument parser error: %s\n", err)
		os.Exit(1)
	}

	conn, cs, err := moteconnection.CreateConnection(opts.ConnectionString)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	dpd, err := director.NewDeviceParameterDirector(conn, opts.Group, opts.Address,
		director.Timeout(time.Duration(opts.Timeout)*time.Second),
		director.Retries(opts.Retries))

	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	logger := logsetup(len(opts.Debug))
	if len(opts.Debug) > 2 {
		conn.SetLoggers(logger)
	}
	dpd.SetLoggers(logger)

	conn.Autoconnect(10 * time.Second)

	time.Sleep(5 * time.Second)

	if len(opts.Quiet) == 0 {
		if conn.Connected() {
			logger.Info.Printf("Connected with %s\n", cs)
		} else {
			logger.Info.Printf("Not (yet?) connected with %s\n", cs)
		}
	}

	success := false

	if len(opts.Template) > 0 && len(opts.List) > 0 {
		err = dpd.StartWithTemplate(opts.Positional.File, opts.Template, opts.List)
	} else {
		err = dpd.Start(opts.Positional.File)
	}
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	} else {
		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt, os.Kill)

		for interrupted := false; interrupted == false; {
			select {
			case sig := <-signals:
				signal.Stop(signals)
				logger.Debug.Printf("signal %s\n", sig)
				interrupted = true
			case <-time.After(time.Second):
				if dpd.Finished() {
					interrupted = true
					success = true
				}
			}
		}

		dpd.Stop()
	}

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
