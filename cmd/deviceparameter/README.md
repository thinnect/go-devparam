deviceparameter(1) -- get/set individual device parameters, get all parameters.
=============================================

## SYNOPSIS

`deviceparameter` ...<br>
`deviceparameter` `-a` _addr_ `-g` _group_ `-d` _dest_ ...<br>
`deviceparameter` `-p` _parameter_ ...<br>
`deviceparameter` `-p` _parameter_ `-v` _value_ ...<br>
`deviceparameter` `-a` _addr_ `-g` _group_ `-d` _dest_ `-p` _parameter_ ...<br>
`deviceparameter` `-a` _addr_ `-g` _group_ `-d` _dest_ `-p` _parameter_ `-v` _value_ ...<br>
`deviceparameter` `--help`<br>

## DESCRIPTION

**deviceparameter** configures or queries device parameters from Mist nodes
using the deviceparameters protocol: <https://github.com/thinnect/tos-devparam>.

In its default mode, `deviceparameter` will sequentially query all parameters
from the locally connected device. It is possible to query individual parameters
and set their values one by one with the `-p` or `--parameter` option.

It is possible to query and configure remote devices by specifying the source
and destination addresses.

Internally the deviceparameters protocol uses length and a byte array to transmit
values. In order to correctly send an integer value, the type of the integer
needs to be known. A typed value parameter can be specified with `--u8`, `--u16`,
`--u32`, `--u64`, `--i8`, `--i16`, `--i32` or `--i64`. The `-v` or `--value`
option will parse the input as a raw hex string, converting it directly to
binary. ASCII strings can be specified with the `--str` option.

The `--timeout` and `--retries` options change how long a single parameter is
tried before skipping to the next one or giving up.

## OPTIONS

Options control connection parameters:

  * `connection`:
  This positional argument is used to specify the connection string for the
  mist network connection. Use sf@HOST:PORT for a SerialForwarder connection or
  serial@PORT:BAUD for a direct serial port.
  The default is sf@localhost:9002.

  * `-g`, `--group`:
  option is used to set the ActiveMessage group. The default is 22,
the value is parsed as a hex string (0x22).

  * `-a`, `--address`:
  option is used to set the source ActiveMessage address used for remote
  requests. The default is 5678, the value is parsed as a hex string (0x5678).

  * `-d`, `--destination`:
  option is used to set the destination ActiveMessage address and switch over to
  a remote request. The value is parsed as a hex string.

Options for controlling task processing timings:

  * `--timeout`:
  The time spent waiting for a response for a configuration action or query.
  Value is in seconds, default is 30.

  * `--retries`:
  The number of attempts made to configure or query a single parameter during
  one operation. The default is 2.

Options for setting the value:

  * `-v`, `--value`:
    The is parsed as hex strings and converted to binary.
    The value must have a length that is divisible by 2 and only contain symbols
    [0-9abcdef]. Value is not case sensitive.
  * `--str`:
    The value is converted to binary using the ASCII encoding.
  * `--u8`:
    The value is converted to an unsigned 8-bit big-endian integer.
  * `--u16`:
    The value is converted to an unsigned 16-bit big-endian integer.
  * `--u32`:
    The value is converted to an unsigned 32-bit big-endian integer.
  * `--u64`:
    The value is converted to an unsigned 64-bit big-endian integer.
  * `--i8`:
    The value is converted to a signed 8-bit big-endian integer.
  * `--i16`:
    The value is converted to a signed 16-bit big-endian integer.
  * `--i32`:
    The value is converted to a signed 32-bit big-endian integer.
  * `--i64`:
    The value is converted to a signed 64-bit big-endian integer.

Miscellaneous options:

  * `-Q`, `--quiet`:
  Turn on quite mode, only parameter values are printed.

  * `-D`, `--debug`:
  Turn on debug mode, can be specified multiple times to increase verbosity.

  * `-V`, `--version`:
  Show the application version.

## EXAMPLES

Query all parameters of the locally connected device:

    $ deviceparameter
    2019/01/28 17:13:36.83 Connected with sf@localhost:9002
    2019/01/28 17:13:36.83 Get parameter list:
    2019/01/28 17:13:36.84  0: tos_node_id 1234
    2019/01/28 17:13:38.86  1: radio_channel 26
    ...
    2019/01/28 17:13:50.052982 21: uptime 18073646
    2019/01/28 17:13:50.210797 Done

Set the radio channel on the locally connected device:

    $ deviceparameter -p radio_channel --u8 25
    2019/01/28 17:16:21.00 Connected with sf@localhost:9002
    2019/01/28 17:16:21.00 Set radio_channel to 0x19
    2019/01/28 17:16:21.01 radio_channel = 25
    2019/01/28 17:16:21.16 Done

Set the name parameter on a remote device:

    $ deviceparameter -a 1234 -d 6789 -p name --str FooBar
    2019/01/28 17:16:22.01 Connected with sf@localhost:9002
    2019/01/28 17:16:22.01 Set name to 0x466F6F426172
    2019/01/28 17:16:32.02 name = FooBar
    2019/01/28 17:16:32.17 Done

## ENVIRONMENT

**deviceparameter** currently does not take any configuration from the environment.

## BUGS

**deviceparameter** is written in go and an issue tracker is available at
<https://github.com/thinnect/go-devparam/issues>.

## COPYRIGHT

**deviceparameter** is Copyright (C) 2019 Thinnect Inc. <http://www.thinnect.com>

## SEE ALSO

deviceparameters(1)
