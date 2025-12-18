module github.com/thinnect/go-devparam/cmd/deviceparameter

go 1.17

replace github.com/thinnect/go-devparam => ../..

require (
	github.com/jessevdk/go-flags v1.5.0
	github.com/proactivity-lab/go-loggers v0.0.0-20180417085828-f892709079bd
	github.com/raidoz/go-moteconnection v0.0.3
	github.com/thinnect/go-devparam v0.0.0-00010101000000-000000000000
)

replace github.com/raidoz/go-moteconnection => /home/raido/workspace_go/src/github.com/raidoz/go-moteconnection

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/joaojeronimo/go-crc16 v0.0.0-20140729130949-59bd0194935e // indirect
	github.com/streadway/amqp v1.1.0 // indirect
	go.bug.st/serial.v1 v0.0.0-20191202182710-24a6610f0541 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
