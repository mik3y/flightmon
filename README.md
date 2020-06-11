# Flightmon

A command-line "GUI" for watching ADS-B output from a dump1090 daemon.

## Usage

```
Usage flightmon:
  -debug
        Enable debug output
  -dump_host string
        dump1090 SBS1 stream address (required) (default "localhost:30003")
  -max_age int
        Stop showing aircraft older than this many seconds (default 60)
  -show_ui
        Enable the UI (default true)
```

Example:

```
$ flightmon --dump_host=localhost:30003
```
