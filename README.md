# Air Quality Server

The Air Quality Server (AQ Server) provides a simple tool for
receiving real time data from the Trondheim Kommune Air Quality
Sensors and storing them to a database.  

The sensors are connected via Narrowband IoT to the [Lab5e Span
service](https://span.lab5e.com/) - a self service, easy-to-use
service to allow developers to get on with their lives when connecting
IoT devices to the 4G network.

Per default the air quality server comes with a built in SQLite
database so that there is no need to manage a separate database
instance for casual use.

## Building

In order to build aq server you need Go version 1.19 or newer.  *It
will probably build with older versions of Go, but we do not support
older versions*.

The first time you build you need to make sure that the required
dependencies are installed.  You can do this simply by issuing the
command

    make dep

Then you can build the project with

    make

*This has been tested on OSX and Linux.  Please let us know if it
works on Windows*.

This will produce a binary `bin/aq`.
