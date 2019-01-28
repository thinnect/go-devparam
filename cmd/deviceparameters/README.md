deviceparameters(1) -- set/query device parameters based on a task list.
=============================================

## SYNOPSIS

`deviceparameters` _file_ ...<br>
`deviceparameters` _file_ `--timeout` _seconds_ `--retries` _count_ ...<br>
`deviceparameters` _file_ `--template` _template_ `--list` _nodelist_ ...<br>
`deviceparameters` `--help`<br>

## DESCRIPTION

**deviceparameters** configures or queries device parameters from Mist nodes
using the deviceparameters protocol: <https://github.com/thinnect/tos-devparam>.

In its default mode, `deviceparameters` takes a task list in csv format and
tries to configure or query the specified parameters on the specified nodes.
`deviceparameters` will update the input file as the procedure progresses,
and the command can be started on the same file several times, if some tasks
were left unfinished. Only tasks that do not have an actual value listed will
be processed.

The task list CSV file must have the following fields:
  * `address`:
    The address specifies the short 16-bit ActiveMessage address.

  * `parameter`:
    The parameter is the parameter name (up to 16 characters).

  * `type`:
    The type is the parameter type, listed in the PARAMETER TYPES section.

  * `desired`:
    The desired value specifies the value that needs to be configured. The value
    is parsed based on the specified type. If the desired value is an empty
    string,then no configuration takes place a query is performed instead. To
    set a parameter to a 0-length empty value, the `nil` type should be used.
    Queries with the `nil` type cannot be performed.

  * `actual`:
    The actual value field is filled by the application, it is used to determine
    whether any action should be taken - if the field is not empty, the task is
    considered complete even if the value does not match the value in the
    desired value field.

  * `info`:
    The info field is filled by the application, it will either contain a
    timestamp if the process was completed or an error message.

`deviceparameters` operates on the task list one node at a time. If it completes
a task successfully, it continues to the next task on the same node. If a task
times out, then `deviceparameters` will move on to the next node.
`deviceparameters` will keep trying failed tasks again once it has gone through
the entire list of nodes. `deviceparameters` will exit only once all tasks
are complete or in failure states that are deemed unresolvable.

The `--timeout` and `--retries` options change how long a single task is tried
before moving to the next one.

Optionally the task list may be automatically generated from a template and
node list, specified with `--template` and `--list` respectively.

## PARAMETER TYPES

The parameter type field is used to determine the method for parsing the desired
value field. It has no effect when the desired value is not specified (a query)
and it will be updated according to the information received from the node.

  * `raw`:
    Values of the raw type are treated as hex strings and converted to binary.
    The value must have a length that is divisible by 2 and only contain symbols
    [0-9abcdef]. Value is not case sensitive.
  * `str`:
    String values are converted to binary using the ASCII encoding. For transmitting
    non-ASCII symbols the `raw` type and manual pre-conversion should be used.
  * `u8`:
    The value is converted to an unsigned 8-bit big-endian integer.
  * `u16`:
    The value is converted to an unsigned 16-bit big-endian integer.
  * `u32`:
    The value is converted to an unsigned 32-bit big-endian integer.
  * `u64`:
    The value is converted to an unsigned 64-bit big-endian integer.
  * `i8`:
    The value is converted to a signed 8-bit big-endian integer.
  * `i16`:
    The value is converted to a signed 16-bit big-endian integer.
  * `i32`:
    The value is converted to a signed 32-bit big-endian integer.
  * `i64`:
    The value is converted to a signed 64-bit big-endian integer.
  * `nil`:
    The nil type can be used to set the length of variable length parameters
    (`raw` and `str`) to 0.

## FILES

The `deviceparameters` command expects a CSV formatted task file as input, with
the header `address, parameter, type, desired value, actual value, info`.

Alternatively a template file and a node list can be specified with the
`--template` and `--list` options. The template and node list are only used if
the task file does not exist, in which case it is generated based on the node
list and template.

The template file follows the same format as the task file, but the address
field is ignored.

The node list is just a list of node addresses with one hexadecimal node
address on each line.

In all files a line beginning with # is considered to be disabled, but must
still conform to the format of their respective file type, free-form comments
are not supported.

## OPTIONS

Options control connection parameters:

  * `-c`, `--conn`:
  The option is used to specify the connection string for the mist network
  connection. Use sf@HOST:PORT for a SerialForwarder connection or
  serial@PORT:BAUD for a direct serial port as.
  The default is sf@localhost:9002.

  * `-g`, `--group`:
  option is used to set the ActiveMessage group. The default is 22,
the value is parsed as a hex string (0x22).

  * `-a`, `--address`:
  option is used to set the source ActiveMessage address.
  The default is 5678, the value is parsed as a hex string (0x5678).

Options for controlling task processing timings:

  * `--timeout`:
  The time spent waiting for a response for a configuration action or query.
  Value is in seconds, default is 30.

  * `--retries`:
  The number of attempts made to configure or query a single parameter during
  one operation. The default is 2.

Task template and node list options:

  * `--template`:
  Path to the task template. See the FILES section for more details.

  * `--list`:
  Path to the node list. See the FILES section for more details.

Miscellaneous options:

  * `-D`, `--debug`:
  Turn on debug mode, can be specified multiple times to increase verbosity.

  * `-V`, `--version`:
  Show the application version.

## EXAMPLES

Execute deviceparameters through sf@localhost.9002 and query the uptime of 2 nodes:

    tasks.csv before:
    address, parameter, type, desired, actual, info
    1234,uptime,u32,,,
    5678,uptime,u32,,,

    Execute deviceparameters:
    $ deviceparameters tasks.csv

    tasks.csv after:
    address, parameter, type, desired, actual, info
    1234,uptime,u32,,123456,2019-01-01T12:00:01Z
    5678,uptime,u32,,235467,2019-01-01T12:00:31Z

Execute deviceparameters through sf@localhost.9002 and set the name of 2 nodes:

    tasks.csv before:
    address, parameter, type, desired, actual, info
    1234,name,str,"node 1",,
    5678,name,str,"node 2",,

    Execute deviceparameters:
    $ deviceparameters tasks.csv

    tasks.csv after:
    address, parameter, type, desired, actual, info
    1234,name,str,"node 1","node 1",2019-01-01T13:00:01Z
    5678,name,str,"node 2","node 2",2019-01-01T13:00:20Z

Execute deviceparameters with additional options:

    $ deviceparameters -a 1234 -g 57 --conn sf@localhost:32000 --retries 3 --timeout 60 tasks.csv --template template.csv --list nodes.txt

## ENVIRONMENT

**deviceparameters** currently does not take any configuration from the environment.

## BUGS

**deviceparameters** is written in go and an issue tracker is available at
<https://github.com/thinnect/go-devparam/issues>.

## COPYRIGHT

**deviceparameters** is Copyright (C) 2019 Thinnect Inc. <http://www.thinnect.com>

## SEE ALSO

deviceparameter(1)
