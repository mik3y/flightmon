# Flightmon

A command-line "GUI" for watching ADS-B output from a dump1090 daemon.

![Screenshot of Flightmon in action](https://user-images.githubusercontent.com/390829/84444664-48db3080-ac10-11ea-994c-3910e9f382fd.png)

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
