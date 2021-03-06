// Author  Raido Pahtma
// License MIT

package deviceparameters

import "fmt"
import "time"
import "errors"
import "bytes"
import "encoding/binary"

import "github.com/proactivity-lab/go-loggers"
import "github.com/proactivity-lab/go-moteconnection"

type DeviceParameter struct {
	Name      string
	Type      DeviceParameterType
	Seqnum    uint8
	Value     []byte
	Timestamp time.Time
	Error     error
}

const TOS_SERIAL_DEVICE_PARAMETERS_ID = 0x80
const AMID_DEVICE_PARAMETERS = 0x82

type DeviceParameterManager struct {
	loggers.DIWEloggers
	sfc moteconnection.MoteConnection
	dsp moteconnection.Dispatcher

	values    map[string]*DeviceParameter
	devstart  time.Time
	heartbeat time.Time

	timeout time.Duration
	retries int

	receive chan moteconnection.Packet

	destination moteconnection.AMAddr // Optional destination

	done   chan bool
	closed bool
}

type ParameterError struct{ s string }
type InvalidParameterValueError struct{ s string }
type ValueMismatchError struct{ s string }
type TimeoutError struct{ s string }

func (self ParameterError) Error() string             { return self.s }
func NewParameterError(text string) error             { return &ParameterError{text} }
func (self InvalidParameterValueError) Error() string { return self.s }
func NewInvalidParameterValueError(text string) error { return &InvalidParameterValueError{text} }
func (self ValueMismatchError) Error() string         { return self.s }
func NewValueMismatchError(text string) error         { return &ValueMismatchError{text} }
func (self TimeoutError) Error() string               { return self.s }
func NewTimeoutError(text string) error               { return &TimeoutError{text} }

func NewDeviceParameterManager(sfc moteconnection.MoteConnection) *DeviceParameterManager {
	dpm := new(DeviceParameterManager)
	dpm.InitLoggers()
	dpm.values = make(map[string]*DeviceParameter)
	dpm.done = make(chan bool)
	dpm.closed = false
	dpm.receive = make(chan moteconnection.Packet)
	dpm.timeout = time.Second
	dpm.retries = 3

	dsp := moteconnection.NewPacketDispatcher(moteconnection.NewRawPacket(TOS_SERIAL_DEVICE_PARAMETERS_ID))
	dsp.RegisterReceiver(dpm.receive)
	dpm.dsp = dsp

	dpm.sfc = sfc
	dpm.sfc.AddDispatcher(dpm.dsp)

	go dpm.run()
	return dpm
}

func NewDeviceParameterActiveMessageManager(sfc moteconnection.MoteConnection, group moteconnection.AMGroup, address moteconnection.AMAddr, destination moteconnection.AMAddr) *DeviceParameterManager {
	dpm := new(DeviceParameterManager)
	dpm.InitLoggers()
	dpm.values = make(map[string]*DeviceParameter)
	dpm.done = make(chan bool)
	dpm.closed = false
	dpm.receive = make(chan moteconnection.Packet)
	dpm.timeout = time.Second
	dpm.retries = 3
	dpm.destination = destination

	dsp := moteconnection.NewMessageDispatcher(moteconnection.NewMessage(group, address))
	dsp.RegisterMessageReceiver(AMID_DEVICE_PARAMETERS, dpm.receive)
	dpm.dsp = dsp

	dpm.sfc = sfc
	dpm.sfc.AddDispatcher(dpm.dsp)

	go dpm.run()
	return dpm
}

func (self *DeviceParameterManager) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
}

func (self *DeviceParameterManager) SetRetries(retries int) {
	self.retries = retries
}

func (self *DeviceParameterManager) GetValue(name string) (*DeviceParameter, error) {
	// Interrupt the run goroutine
	self.done <- true

	var result error = errors.New("disabled")

	for retries := 0; retries <= self.retries; retries++ {
		// Send get request
		msg := self.dsp.NewPacket()
		if self.destination != 0 {
			msg.(*moteconnection.Message).SetDestination(self.destination)
			msg.(*moteconnection.Message).SetType(AMID_DEVICE_PARAMETERS)
		}
		payload := new(DpGetParameterId)
		payload.Header = DP_GET_PARAMETER_WITH_ID
		payload.Id = name
		msg.SetPayload(moteconnection.SerializePacket(payload))
		self.sfc.Send(msg)

		// Wait for value
		dp, err := self.waitValueId(name)
		if err == nil {
			go self.run()
			return dp, nil
		} else {
			result = err
			if _, ok := err.(*ParameterError); ok {
				break
			}
		}
	}

	go self.run()
	return nil, result
}

func (self *DeviceParameterManager) SetValue(name string, value []byte) (*DeviceParameter, error) {
	// Interrupt the run goroutine
	self.done <- true

	var result error = errors.New("disabled")

	for retries := 0; retries <= self.retries; retries++ {
		// Send set request
		msg := self.dsp.NewPacket()
		if self.destination != 0 {
			msg.(*moteconnection.Message).SetDestination(self.destination)
			msg.(*moteconnection.Message).SetType(AMID_DEVICE_PARAMETERS)
		}
		payload := new(DpSetParameterId)
		payload.Header = DP_SET_PARAMETER_WITH_ID
		payload.Id = name
		payload.Value = value
		msg.SetPayload(moteconnection.SerializePacket(payload))
		self.sfc.Send(msg)

		// Wait for value
		dp, err := self.waitValueId(name)
		if err == nil {
			if bytes.Compare(dp.Value, value) == 0 {
				// store in values table
				self.values[name] = dp
				go self.run()
				return dp, nil
			} else {
				go self.run()
				return dp, NewValueMismatchError(fmt.Sprintf("Returned value %X does not match set value %X!", dp.Value, value))
			}
		} else {
			result = err
			if _, ok := err.(*ParameterError); ok {
				break
			}
		}
	}

	go self.run()
	return nil, result
}

func (self *DeviceParameterManager) GetList() (chan *DeviceParameter, error) {
	// Interrupt the run goroutine
	self.done <- true

	delivery := make(chan *DeviceParameter)
	go self.getList(delivery)

	return delivery, nil
}

func (self *DeviceParameterManager) receivedPacket(msg moteconnection.Packet) {
	self.Debug.Printf("%s\n", msg)
	payload := msg.GetPayload()
	if len(payload) > 0 {
		if payload[0] == DP_HEARTBEAT {
			p := new(DpHeartbeat)
			if err := moteconnection.DeserializePacket(p, payload); err == nil {
				self.heartbeat = time.Now()
				self.devstart = self.heartbeat.Add(-time.Duration(p.Uptime) * time.Second)
				// TODO check stuff
			}
		}
	}
}

func (self *DeviceParameterManager) waitValueId(name string) (*DeviceParameter, error) {
	start := time.Now()
	for {
		select {
		case packet := <-self.receive:
			payload := packet.GetPayload()

			if self.destination != 0 {
				msg, ok := packet.(*moteconnection.Message)
				if !ok || msg.Source() != self.destination {
					self.Debug.Printf("Ignoring packet %s\n", packet)
					payload = nil
				}
			}

			if len(payload) > 0 {
				if payload[0] == DP_PARAMETER {
					p := new(DpParameter)
					if err := moteconnection.DeserializePacket(p, payload); err == nil {
						if p.Id == name {
							return &DeviceParameter{name, DeviceParameterType(p.Type), p.Seqnum, p.Value, time.Now(), nil}, nil
						}
					} else {
						self.Error.Printf("Deserialize error %s %s\n", err, packet)
					}
				} else if payload[0] == DP_ERROR_PARAMETER_ID {
					p := new(DpErrorParameterId)
					if err := moteconnection.DeserializePacket(p, payload); err == nil {
						if p.Id == name {
							if p.Exists {
								if p.Err == 6 { // EINVAL
									return nil, NewInvalidParameterValueError(fmt.Sprintf("Something went wrong with parameter \"%s\", error %d - EINVAL!", name, p.Err))
								}
								return nil, errors.New(fmt.Sprintf("Something went wrong with parameter \"%s\", error %d!", name, p.Err))
							} else {
								return nil, NewParameterError(fmt.Sprintf("No parameter \"%s\" on device!", name))
							}
						} else {
							self.Warning.Printf("Received unexpected error for parameter %s\n", p.Id)
						}
					} else {
						self.Error.Printf("Deserialize error %s %s\n", err, packet)
					}
				} else {
					self.receivedPacket(packet)
				}
			}
		case <-time.After(remaining(start, self.timeout)):
			return nil, NewTimeoutError(fmt.Sprintf("Timeout for parameter \"%s\"!", name))
		}
	}
}

func (self *DeviceParameterManager) waitValueSeqnum(seqnum uint8) (*DeviceParameter, error) {
	start := time.Now()
	for {
		select {
		case packet := <-self.receive:
			payload := packet.GetPayload()

			if self.destination != 0 {
				msg, ok := packet.(*moteconnection.Message)
				if !ok || msg.Source() != self.destination {
					self.Debug.Printf("Ignoring packet %s\n", packet)
					payload = nil
				}
			}

			if len(payload) > 0 {
				if payload[0] == DP_PARAMETER {
					p := new(DpParameter)
					if err := moteconnection.DeserializePacket(p, payload); err == nil {
						if p.Seqnum == seqnum {
							return &DeviceParameter{p.Id, DeviceParameterType(p.Type), p.Seqnum, p.Value, time.Now(), nil}, nil
						}
					} else {
						self.Error.Printf("Deserialize error %s %s\n", err, packet)
					}
				} else if payload[0] == DP_ERROR_PARAMETER_SEQNUM {
					p := new(DpErrorParameterSeqnum)
					if err := moteconnection.DeserializePacket(p, payload); err == nil {
						if p.Seqnum == seqnum {
							if p.Exists {
								return nil, errors.New(fmt.Sprintf("Something went wrong with parameter %d, error %d!", seqnum, p.Err))
							} else {
								return nil, NewParameterError(fmt.Sprintf("No parameter %d on device!", seqnum))
							}
						} else {
							self.Warning.Printf("Received unexpected error for parameter %d\n", p.Seqnum)
						}
					} else {
						self.Error.Printf("Deserialize error %s %s\n", err, packet)
					}
				} else {
					self.receivedPacket(packet)
				}
			}
		case <-time.After(remaining(start, self.timeout)):
			return nil, NewTimeoutError(fmt.Sprintf("Timeout for parameter %d!", seqnum))
		}
	}
}

func (self *DeviceParameterManager) getList(delivery chan *DeviceParameter) {
	for i := 0; i < 256; i++ {
		for retries := 0; retries <= self.retries; retries++ {
			self.Debug.Printf("Get %d %d/%d\n", i, retries, self.retries)
			// Send get request
			msg := self.dsp.NewPacket()
			if self.destination != 0 {
				msg.(*moteconnection.Message).SetDestination(self.destination)
				msg.(*moteconnection.Message).SetType(AMID_DEVICE_PARAMETERS)
			}
			payload := new(DpGetParameterSeqnum)
			payload.Header = DP_GET_PARAMETER_WITH_SEQNUM
			payload.Seqnum = uint8(i)
			msg.SetPayload(moteconnection.SerializePacket(payload))
			self.sfc.Send(msg)

			// Wait for value
			dp, err := self.waitValueSeqnum(uint8(i))
			if err == nil {
				delivery <- dp
				break
			} else {
				self.Debug.Printf("Got %s\n", err)
				if _, ok := err.(*ParameterError); ok { // This parameter does not exist and therefore the list is complete
					self.Debug.Printf("closing")
					close(delivery)
					go self.run()
					return
				} else if retries == self.retries {
					delivery <- &DeviceParameter{"", 0, uint8(i), nil, time.Now(), err}
					break
				}
			}
		}
	}

	close(delivery)
	go self.run()
}

func (self *DeviceParameterManager) run() {
	self.Debug.Printf("DPM running\n")
	for {
		select {
		case packet := <-self.receive:
			msg := packet
			self.receivedPacket(msg)
		case done := <-self.done:
			if done {
				self.Debug.Printf("DPM interrupted\n")
			} else {
				self.Debug.Printf("DPM closed\n")
			}
			return
		}
	}
}

func (self *DeviceParameterManager) Close() error {
	if !self.closed {
		self.closed = true
		close(self.done)
		self.sfc.RemoveDispatcher(self.dsp)
		return nil
	}
	return errors.New("Close has already been called!")
}

func (self *DeviceParameter) String() string {
	if self.Type == DP_TYPE_RAW {
		return fmt.Sprintf("%X", self.Value)
	} else if self.Type == DP_TYPE_STRING {
		return string(self.Value)
	} else {
		s := fmt.Sprintf("%v", self.Value)
		buf := bytes.NewBuffer(self.Value)
		if self.Type == DP_TYPE_UINT8 {
			var v uint8
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_UINT16 {
			var v uint16
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_UINT32 {
			var v uint32
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_UINT64 {
			var v uint64
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_INT8 {
			var v int8
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_INT16 {
			var v int16
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_INT32 {
			var v int32
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		} else if self.Type == DP_TYPE_INT64 {
			var v int64
			if err := binary.Read(buf, binary.BigEndian, &v); err == nil {
				s = fmt.Sprintf("%d", v)
			}
		}
		return s
	}
}

func remaining(start time.Time, timeout time.Duration) time.Duration {
	elapsed := time.Since(start)
	if elapsed < timeout {
		return timeout - elapsed
	}
	return 0
}
