// Author  Raido Pahtma
// License MIT

package director

import "os"
import "io"
import "bufio"

import "fmt"
import "time"
import "strconv"
import "strings"

import "errors"

import "encoding/csv"

import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-moteconnection"

import dp "github.com/thinnect/go-devparam"

type DeviceParameterTask struct {
	Address   moteconnection.AMAddr
	Parameter string
	Type      dp.DeviceParameterType
	Desired   []byte
	Actual    []byte
	Info      string

	Disabled bool // Has been commented out
	Blocked  bool // Something wrong with it
}

type DeviceParameterDirector struct {
	loggers.DIWEloggers

	conn    moteconnection.MoteConnection
	group   moteconnection.AMGroup
	address moteconnection.AMAddr
	//dsp  moteconnection.Dispatcher

	timeout time.Duration
	retries uint8

	filepath string

	tasks []DeviceParameterTask

	interrupt chan bool
	done      chan bool
}

type option func(*DeviceParameterDirector) (option, error)

func (dpd *DeviceParameterDirector) Option(opts ...option) (option, error) {
	var prev option
	for _, opt := range opts {
		previous, err := opt(dpd)
		if err != nil {
			return previous, err
		}
		prev = previous
	}
	return prev, nil
}

func NewDeviceParameterDirector(conn moteconnection.MoteConnection,
	group moteconnection.AMGroup, address moteconnection.AMAddr,
	opts ...option) (*DeviceParameterDirector, error) {

	dpd := new(DeviceParameterDirector)
	dpd.conn = conn
	dpd.group = group
	dpd.address = address

	dpd.timeout = 30 * time.Second
	dpd.retries = 2

	dpd.interrupt = make(chan bool)

	for _, opt := range opts {
		_, err := opt(dpd)
		if err != nil {
			return nil, err
		}
	}

	return dpd, nil
}

func Timeout(t time.Duration) option {
	return func(dpd *DeviceParameterDirector) (option, error) {
		previous := dpd.timeout
		dpd.timeout = t
		return Timeout(previous), nil
	}
}

func Retries(r uint8) option {
	return func(dpd *DeviceParameterDirector) (option, error) {
		previous := dpd.retries
		dpd.retries = r
		return Retries(previous), nil
	}
}

func (task *DeviceParameterTask) ToCSV() []string {
	addr := task.Address.String()
	dv := ""
	if task.Desired != nil {
		dv, _ = dp.ParameterValueString(task.Type, task.Desired)
	}
	av := ""
	if task.Actual != nil {
		av, _ = dp.ParameterValueString(task.Type, task.Actual)
	}
	if task.Disabled {
		addr = "#" + addr
	}
	return []string{addr, task.Parameter, task.Type.String(), dv, av, task.Info}
}

func (dpd *DeviceParameterDirector) writeTasksToFile(tasks []DeviceParameterTask, filepath string) error {
	file, err := os.Create(filepath)
	if err == nil {
		defer file.Close()

		w := csv.NewWriter(file)
		defer w.Flush()

		if err := w.Write([]string{"address", "parameter", "type", "desired", "actual", "info"}); err != nil {
			dpd.Error.Printf("error writing header: %s", err)
			return err
		}

		for _, task := range tasks {
			if err := w.Write(task.ToCSV()); err != nil {
				dpd.Error.Printf("error writing output: %s", err)
				return err
			}
		}

		return nil
	}
	return err
}

func (dpd *DeviceParameterDirector) updateOutput() {
	newfile := dpd.filepath + ".new"
	if err := dpd.writeTasksToFile(dpd.tasks, newfile); err == nil {
		err = os.Rename(newfile, dpd.filepath)
		if err != nil {
			dpd.Error.Printf("error updating file: %s", err)
		}
	} else {
		dpd.Error.Printf("error updating file: %s", err)
	}
}

func (dpd *DeviceParameterDirector) run() {
	dpd.Debug.Printf("%d tasks in queue\n", len(dpd.tasks))

	interrupted := false
	for interrupted == false {
		// organize a queue of nodes
		ns := make(map[moteconnection.AMAddr]bool)
		for _, task := range dpd.tasks {
			if task.Disabled == false && task.Blocked == false && task.Actual == nil {
				ns[task.Address] = true
			}
		}
		if len(ns) == 0 {
			break
		}
		q := make([]moteconnection.AMAddr, 0, len(ns))
		for k := range ns {
			q = append(q, k)
		}

		dpd.Debug.Printf("%d nodes in queue\n", len(q))
		// start processing the queue
		for _, node := range q {
			dpm := dp.NewDeviceParameterActiveMessageManager(dpd.conn, dpd.group, dpd.address, node)
			dpm.SetTimeout(dpd.timeout)
			dpm.SetRetries(int(dpd.retries))

			for idx, task := range dpd.tasks { // look for a suitable task
				if task.Disabled == false && task.Blocked == false && task.Address == node && task.Actual == nil {
					dpd.Debug.Printf("%+v\n", task)
					skip := false
					if task.Desired == nil && task.Type != dp.DP_TYPE_NIL { // only a read is requested
						if val, err := dpm.GetValue(task.Parameter); err == nil {
							task.Type = val.Type
							task.Actual = val.Value
							task.Info = time.Now().UTC().Format("2006-01-02T15:04:05Z")
							dpd.Info.Printf("Got parameter %s from node %s.\n", task.Parameter, task.Address)
						} else {
							dpd.Warning.Printf("Failed to get parameter %s from node %s.\n", task.Parameter, task.Address)
							task.Info = err.Error()
							switch err.(type) {
							case dp.ParameterError: // No such parameter?
								task.Blocked = true
							case dp.TimeoutError:
								// just keep trying, but skip to the next node
								skip = true
							default:
								skip = true
							}
						}
					} else { // must set value
						if val, err := dpm.SetValue(task.Parameter, task.Desired); err == nil {
							if task.Type != val.Type {
								dpd.Warning.Printf("Parameter %s set on node %s, but types did not match: %s / %s\n", task.Parameter, task.Address, task.Type, val.Type)
							}
							task.Type = val.Type
							task.Actual = val.Value
							task.Info = time.Now().UTC().Format("2006-01-02T15:04:05Z")
							dpd.Info.Printf("Set parameter %s on node %s.\n", task.Parameter, task.Address)
						} else {
							dpd.Warning.Printf("Failed to set parameter %s on node %s, result=%s.\n", task.Parameter, task.Address, err.Error())
							task.Info = err.Error()
							switch err.(type) {
							case dp.InvalidParameterValueError: // The type is probably bad
								task.Blocked = true
							case dp.ParameterError: // No such parameter?
								task.Blocked = true
							case dp.ValueMismatchError:
								task.Actual = val.Value
								// not updating the type here just yet
								task.Blocked = true // blocking it until more advanced handling is added
							case dp.TimeoutError: // just keep trying, but skip to the next node
								skip = true
							default: // Unexpected, possibly EBUSY?
								skip = true
							}
						}
					}

					dpd.tasks[idx] = task
					dpd.updateOutput()

					if skip {
						break // proceed to next node in the queue
					}
				}

				select {
				case <-dpd.interrupt:
					dpd.Debug.Println("interrupted")
					interrupted = true
				default:
				}
				if interrupted {
					break
				}
			}

			dpm.Close() // de-initialize the manager, since manager is target specific and moving to next one

			select {
			case <-dpd.interrupt:
				dpd.Debug.Println("interrupted")
				interrupted = true
			default:
			}
			if interrupted {
				break
			}
		}
	}

	close(dpd.done)
}

func (dpd *DeviceParameterDirector) readTaskFile(filepath string) ([]DeviceParameterTask, error) {
	tasks := make([]DeviceParameterTask, 0)

	csvf, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer csvf.Close()

	reader := csv.NewReader(bufio.NewReader(csvf))
	//reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = 6
	//reader.Comment = '#' // Want to preserve comments

	for l := 0; true; l++ {
		line, err := reader.Read()
		if err == io.EOF {
			dpd.Debug.Printf("Read %d lines from %s.\n", l, filepath)
			break
		} else if err != nil {
			dpd.Error.Printf("%s", err)
			return nil, err
		}

		if line[0] == "address" {
			continue // found the header
		}

		var task DeviceParameterTask

		// Tasks/lines in the CSV can be disabled by putting a # in front of them
		if strings.HasPrefix(line[0], "#") {
			task.Disabled = true
			line[0] = line[0][1:len(line[0])]
		}

		// validate node address
		addr64, err := strconv.ParseUint(line[0], 16, 16)
		if err != nil {
			return nil, err
		}
		addr := moteconnection.AMAddr(addr64)

		if 0 < addr && addr < 0xFFFF {
			task.Address = addr
		} else {
			return nil, errors.New(fmt.Sprintf("'%s' is not a valid address!", line[0]))
		}
		// validate parameter name
		if 0 < len(line[1]) && len(line[1]) <= 16 {
			task.Parameter = line[1]
		} else {
			return nil, errors.New(fmt.Sprintf("'%s' is not a valid parameter name!", line[1]))
		}
		// validate parameter type
		task.Type, err = dp.ParseDeviceParameterType(line[2])
		if err != nil {
			return nil, err
		}
		// validate parameter desired value
		if len(line[3]) == 0 {
			task.Desired = nil
		} else {
			task.Desired, err = dp.ParseParameterValue(task.Type, line[3])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("'%s' is not a valid parameter value!", line[3]))
			}
		}
		// validate parameter actual value field
		if len(line[4]) == 0 {
			task.Actual = nil
		} else {
			task.Actual, err = dp.ParseParameterValue(task.Type, line[4])
			if err != nil {
				return nil, errors.New(fmt.Sprintf("'%s' is not a valid parameter value!", line[4]))
			}
		}
		// validate the timestamp?
		task.Info = line[5]

		// dpd.Debug.Printf("%+v\n", task)

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (dpd *DeviceParameterDirector) readNodeFile(filepath string) ([]moteconnection.AMAddr, error) {
	nf, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer nf.Close()

	scanner := bufio.NewScanner(bufio.NewReader(nf))

	nodes := make([]moteconnection.AMAddr, 0)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if len(t) > 0 && strings.HasPrefix(t, "#") == false {
			if addr, err := strconv.ParseUint(t, 16, 16); err == nil {
				nodes = append(nodes, moteconnection.AMAddr(addr))
			} else {
				return nil, err
			}
		}
	}
	return nodes, nil
}

func (dpd *DeviceParameterDirector) Start(filepath string) error {
	// open and validate the file
	dpd.filepath = filepath

	tasks, err := dpd.readTaskFile(filepath)
	if err != nil {
		return err
	}
	dpd.tasks = tasks

	// setup sniffing dispatchers
	// generate statistics for choosing next target?

	// start statemachine
	dpd.done = make(chan bool)
	go dpd.run()

	return nil
}

func (dpd *DeviceParameterDirector) StartWithTemplate(filepath string, template string, nodelist string) error {
	if _, err := os.Stat(filepath); err == nil {
		dpd.Info.Printf("Task file exists, not using template.")
	} else {
		templateTasks, err := dpd.readTaskFile(template)
		if err != nil {
			return err
		}

		nodes, err := dpd.readNodeFile(nodelist)
		if err != nil {
			return err
		}

		tasks := make([]DeviceParameterTask, 0, len(nodes)*len(templateTasks))
		for _, node := range nodes {
			for _, task := range templateTasks {
				task.Address = node
				tasks = append(tasks, task)
			}
		}

		dpd.writeTasksToFile(tasks, filepath)
	}
	return dpd.Start(filepath)
}

func (dpd *DeviceParameterDirector) Finished() bool {
	select {
	case <-dpd.done:
		return true
	default:
	}
	return false
}

func (dpd *DeviceParameterDirector) Stop() {
	close(dpd.interrupt) // interrupt the statemachine
	<-dpd.done           // wait for it to finish
	// deinitialize the dispatchers
}
