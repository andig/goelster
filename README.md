`goelster` is a CAN bus utility for interacting with Elster/Kromschröder CAN bus devices. Popular heat pumps like Stiebel Eltron or WPM Wärmepumpenmanager use this protocol.

`goelster` is based on the work of [Jürg Müller](http://juerg5524.ch/list_data.php) and his `can_progs` package. This tool would not have been possible without the work and efforts of Jürg.

# Prerequisites

`goelster` requires Go 1.11 for compiling. Earlier versions of go can be used if required modules are installed. 
Go can be downloaded from [golang.org/dl/](https://golang.org/dl/). For Raspbian chose `linux-armv6l.tar.gz` and move to `/usr/local`.

# Usage

`goelster` command line is similar to the `can_progs` package.

## Listening on the CAN bus

Listening on the CAN bus is similar to the `can_logger` tool from the `can_progs` package. Listing possible with the following command:

    goelster <can dev>

## Scanning a device

For scanning, `goelster` will try to read every single elster register. For details on all defined readings see Elster reading definitions source [github](https://github.com/andig/goelster/blob/master/readings.go):

    goelster <can dev> <sender can id>

## Reading a device register

    goelster <can dev> <sender can id>.<receiver can id>

The value will be decoded as defined in the [Elster reading definitions](https://github.com/andig/goelster/blob/master/readings.go).

## Writing a device register

Writing supports two modes. For compatibility with `can_scan` it is possible to write **raw binary** values:

    goelster <can dev> <sender can id>.<receiver can id>.<raw value>

Example: set `EINSTELL_SPEICHERSOLLTEMP2` to 42°C

    goelster slcan0 680 180.0a06.01a4

It is also possible to specify numeric values:

    goelster <can dev> <sender can id>.<receiver can id> <value>

Example: set `EINSTELL_SPEICHERSOLLTEMP2` to 42°C

    goelster slcan0 680 180.0a06 42

The value will be encoded as defined in the [Elster reading definitions](https://github.com/andig/goelster/blob/master/readings.go).
