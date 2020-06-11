# Flightmon

A command-line "GUI" for watching ADS-B output from a dump1090 daemon.

![Screenshot of Flightmon in action](https://user-images.githubusercontent.com/390829/84444664-48db3080-ac10-11ea-994c-3910e9f382fd.png)

Test on Mac OS X, Linux amd64, and Linux arm (e.g. Raspberry Pi).

## Usage

To run it, just type `flightmon`. Your terminal will be taken over with a lovely table based on the SBS1 data stream coming from `localhost:30003`. You can specify a different host with `--dump_host`.

Full usage instructions:

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

## Download

Head on over to the [releases page](https://github.com/mik3y/flightmon/releases) for pre-built binaries for Linux and Mac.

## Changelog

See [CHANGELOG.md](https://github.com/mik3y/flightmon/blob/master/CHANGELOG.md) for the latest changes.

## Contributing

Contributions welcome! Open an issue or a pull request if you've got ideas.
