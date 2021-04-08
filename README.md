# flightlog2kml

## Overview

A suite of tools to generate annotated KML/KMZ files (and other data) from **inav** blackbox logs and OpenTX log files (inav S.Port telemetry). From 0.9.7, there is limited support for OpenTX logs from Ardupilot.

* [flightlog2kml](#flightlog2kml) - Generates KML/Z file(s) from Blackbox log(s), OpenTX (OTX) and Bullet GCSS logs
* [fl2mqtt](#fl2mqtt) - Generates MQTT data to stimulate the on-line Ground Control Station [BulletGCSS](https://bulletgcss.fpvsampa.com/)
* fl2ltm - If `fl2mqtt` is installed (typically by hard or soft link) as `fl2ltm` it generates LTM  (inav's Lightweight Telemetry). This is primarily for use by [mwp](https://github.com/stronnag/mwptools/) as a unified replay tool for Blackbox and OpenTx logs.
* [log2mission](#log2mission) - Converts a flight log (Blackbox, OpenTx, BulletGCSS) into a valid inav mission. A number of filters may be applied (time, flight mode).
* [mission2kml](#mission2kml) - Generate KML file from inav mission files (and other formats)

## flightlog2kml

```
$ flightlog2kml --help
Usage of flightlog2kml [options] file...
  -dms
    	Show positions as DD:MM:SS.s (vice decimal degrees) (default true)
  -dump
    	Dump log headers and exit
  -efficiency
    	Include efficiency layer in KML/Z (default true)
  -extrude
    	Extends track points to ground (default true)
  -gradient string
    	Specific colour gradient [red,rdgn,yor] (default "yor")
  -home-alt int
    	[OTX] home altitude
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -kml
    	Generate KML (vice default KMZ)
  -mission string
    	Optional mission file name
  -outdir string
    	Output directory for generated KML
  -rebase string
    	rebase all positions on lat,lon[,alt]
  -rssi
    	Set RSSI view as default
  -split-time int
    	[OTX] Time(s) determining log split, 0 disables (default 120)
  -visibility int
    	0=folder value,-1=don't set,1=all on

flightlog2kml 0.9.7, commit: bccf72e/2021-03-20
```

Multiple logs (with multiple indices) may be given. A KML/Z will be generated for each file / index.

The output file is named from the base name of the source log file, appended with the index number and `.kml` or `.kmz` as appropriate. For example:

```
$ flightlog2kml LOG00044.TXT
Log      : LOG00044.TXT / 1
Flight   : "Model" on 2020-04-12T14:24:01.410+03:00
Firmware : INAV 2.4.0 (bcd4caef9) MATEKF722 of Feb 11 2020 22:48:59
Size     : 19.36 MB
Altitude : 292.8 m at 25:42
Speed    : 28.0 m/s at 13:54
Range    : 17322 m at 14:22
Current  : 30.6 A at 00:05
Distance : 48437 m
Duration : 43:44
Disarm   : Switch

```
results in the KMZ file "LOG00044.1.kmz"

Where `-mission <file>` is given, the given waypoint `<mission file>` will be included in the generated KML/Z; mission files may be one of the following formats as supported by [impload](https://github.com/stronnag/impload):

* MultiWii / XML mission files (MW-XML) ([mwp](https://github.com/stronnag/mwptools/), [inav-configurator](https://github.com/iNavFlight/inav-configurator), [ezgui](https://play.google.com/store/apps/details?id=com.ezio.multiwii&hl=en_GB), [mission planner for inav](https://play.google.com/store/apps/details?id=com.eziosoft.ezgui.inav&hl=en), drone-helper).
* [mwp JSON files](https://github.com/stronnag/mwptools/)
* [apmplanner2](https://ardupilot.org/planner2/) "QGC WPL 110" text files
* [qgroundcontrol](http://qgroundcontrol.com/) JSON plan files
* GPX and CSV (as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide))

If you use a format other than MW-XML or mwp JSON, it is recommended that you review any relevant format constraints as described in the [impload user guide](https://github.com/stronnag/impload/wiki/impload-User-Guide).

### Output

KML/Z file defining tracks which may be displayed Google Earth. Tracks can be animated with the time slider.

Both Flight Mode and RSSI tracks are generated; the default for display is Flight Mode, unless `-rssi` is specified (and RSSI data is available in the log). The log summary is displayed by double clicking on the `file name` folder in Google Earth.

### Modes

`flightlog2kml` can generate three distinct colour-coded outputs:

* Flight mode: the default, colours as [below](#flight_mode_track).
* RSSI mode: RSSI percentage as a colour gradient, according to the current `--gradient` setting. Note that if no valid RSSI is found in the log, this mode will be suppressed.
* Efficiency mode: The efficiency (mAh/km) as a colour gradient,  according to the current `--gradient` setting. This is not enabled by default, and requires the `--efficiency` setting to be specified, either as a command line option or permanently in the [configuration file](#setting-default-options).

#### Flight Mode Track

* White : WP Mission
* Yellow : RTH
* Green : Pos Hold
* Lighter Green : Alt Hold
* Purple : Cruise
* Cyan : Piloted
* Lighter cyan : Launch
* Red : Failsafe
* Orange : Emergency Landing

### Colour Gradients

The RSSI and Efficiency modes are displayed using a colour gradient. Three gradients are available:
* `red` : The default, white representing the best (100%), red the worst (0%)
* `rdgn` : Red to green, green representing the best (100%), red the worst (0%)
* `yor` : Yellow/Orange/Red, yellow representing the best (100%), red the worst (0%)

If no option is given, `red` is assumed. Values are set by the `--gradient` command line option or  in the [configuration file](#setting-default-options).

### Examples

Note: These images are rather old, it looks much better now.

#### Flight Modes

![Example 1](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-1.png)

![Example 2](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-2.png)

![Example 3](https://github.com/stronnag/mwptools/wiki/images/bbl2kml-3.png)

#### RSSI

![Example 4](https://github.com/stronnag/mwptools/wiki/images/inav-tracer-rssi.jpg)

### Using OpenTX logs

There are a few issues with OpenTX logs, the first of which needs OpenTX 2.3.11 (released 2021-01-08) to be resolved:
* CRSF logs in OpenTX 2.3.10 do not record the FM (Flight Mode) field. This makes it impossible to determine flight mode, or even if the craft is armed. Currently `flightlog2kml` tries to evince the armed state from other data.
* GPS Elevation. Unless you have a GPS attached to the TX, you don't get GPS altitude. This can be set by the `-home-alt H` value (in metres). Otherwise `flightlog2kml` will use an online elevation service.
* OpenTX creates a log per calendar day (IIRC), this means there may be multiple logs in the same file. Delimiting these individual logs is less than trivial, to some degree due to the prior CRSF issue which means arm / disarm is not reliably available. Currently, `flightlog2kml` assumes that a gap of more than 120 seconds indicates a new flight. The `-split-time` value allows a user-defined split time (seconds). Setting this to zero disables the log splitting function.

## fl2mqtt

The MQTT option (for BulletGCSS) uses a MQTT broker URI, which may include a username/password and cafile if required for authentication and/or encryption. It can also generate compatible log files that may be replayed by BulletGCSS' internal log player (without requiring a MQTT broker).

```
$ fl2mqtt --help
Usage of fl2mqtt [options] file...
  -blt-vers int
    	[MQTT] BulletGCSS version (default 2)
  -broker string
    	Mqtt URI (mqtt://[user[:pass]@]broker[:port]/topic[?cafile=file]
  -dump
    	Dump log headers and exit
  -home-alt int
    	[OTX] home altitude
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -logfile string
    	Log file for browser replay
  -mission string
    	Optional mission file name
  -rebase string
    	rebase all positions on lat,lon[,alt]
  -split-time int
    	[OTX] Time(s) determining log split, 0 disables (default 120)
```

The [BulletGCSS wiki](https://github.com/danarrib/BulletGCSS/wiki) describes how the broker values are chosen; in general:

* It is safe to use `broker.emqx.io` as the MQTT broker, this is default if no broker host is defined in the URI.
* You should use a unique topic for publishing your own data, this is slash separated string, for example `foo/bar/quux/demo`; the topic should include at least three elements.
* If you want to use a TLS (encrypted) connection to the broker, you may need to supply the broker's CA CRT (PEM) file. A reputable test broker will provide this via their web site.

Note that the scheme (**mqtt**:// in the `--help` text) is interpreted as:

* ws - Websocket (vice TCP socket), ensure the websocket port is also specificed
* wss - Encrypted websocket, ensure the TLS websocket port is also specificed. TLS validation is performed using the operating system.
* mqtts,ssl - Secure (TLS) TCP connection. Ensure the TLS port is specified. TLS validation is performed using the operating system, unless `?cafile=file` is specified.
* mqtt (or any-other scheme) - TCP connection. If `?cafile=file` is specified, then that is used for TLS validation (and the TLS port should be specified).

There is a [bb2kml wiki article](https://github.com/stronnag/bbl2kml/wiki/Self-Hosting-a-MQTT-server-(e.g.-for-fl2mqtt-&--BulletGCSS)) describing how to host your own MQTT broker, for reasons of convenience of better privacy.

Example:
```
## the default broker is used ##
$ fl2mqtt -broker mqtt://broker.emqx.io/org/mwptools/mqtt/playotx openTXlog.csv
$ fl2mqtt -broker mqtt:///org/mwptools/mqtt/playbbl blackbox.TXT

## broker is test.mosquitto.org, over TLS, needs cafile with self-signed certificate
## note the TLS port is also given (8883 in this case)
$ fl2mqtt -broker mqtt://test.mosquitto.org:8883/fl2mqtt/fl2mtqq/test?cafile=mosquitto.org.crt -mission simple_jump.mission BBL_102629.TXT
## No cafile needed, validated certificate
$ fl2mqtt -broker mqtts://broker.emqx.io:8883/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
## Web sockets (plain text / TLS); mosquitto:8081 has valid Lets Encrypt cert.
$ fl2mqtt -broker ws://test.mosquitto.org:8080/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
$ fl2mqtt -broker wss://test.mosquitto.org:8081/fl2mqtt/fl2mtqq/test -mission simple_jump.mission BBL_102629.TXT
```

If a mission file is given, this will also be displayed by BulletGCSS, albeit incorrectly if there WP contains types other than `WAYPOINT` and `RTH`.

[mwp](https://github.com/stronnag/mwptools) can also process / display the BulletGCSS MQTT protocol, using a similar [URI definition](https://github.com/stronnag/mwptools/wiki/mqtt---bulletgcss-telemetry).

## log2mission

`log2mission` will create an inav XML mission file from a supported flight log (Blackbox, OpenTX, BulletGCSS). The mission will not exceed the inav maximum of 60 mission points.

```
$ log2mission
Usage of log2mission [options] file...
  -end-offset int
    	End Offset (seconds) (default -30)
  -epsilon float
    	Epsilon (default 0.015)
  -index int
    	Log index
  -interval int
    	Sampling Interval (ms) (default 1000)
  -mode-filter string
    	Mode filter (cruise,wp)
  -rebase string
    	rebase all positions on lat,lon[,alt]
  -split-time int
    	[OTX] Time(s) determining log split, 0 disables (default 120)
  -start-offset int
    	Start Offset (seconds) (default 30)
```

* The `start-offset` and `end-offset` compensate for the fact that the start / end of the flight is usually on the ground, and thus is not a good WP choice. The defaults are 30 seconds for the start offset and -30 seconds (i.e. 30 seconds from the end) for the end offset. The end offset may be specified as either a positive number of seconds from the start of the log or a negative number (from the end). Locations prior to the start offset and after the end offset are not considered for mission generation. If the `end-offset` is specified (0 cancels it), and there is no flight mode filter, then RTH is included in the generated mission.
* The `mode-filter` allows the log to filtered on Cruise and WP modes, e.g. `-mode-filter=cruise`, `-mode-filter=wp`, `-mode-filter=cruise,wp`. If `mode-filter` is specified, log entries not in the required flight mode(s) are discarded. Cruise includes both 2D and 3D cruise.

### `epsilon` tuning

The `epsilon` value is an opaque factor that controls the point simplification process (using the Ramer–Douglas–Peucker algorithm). The default value should be a good starting point for fixed wing with reasonably sedate flying. On a multi-rotor in a small flight area, a much smaller value (e.g. 0.001) would be more appropriate.  Increasing the value will decrease the number of mission points generated. `log2mission` will do this automatically if the default value results in greater than 60 mission points, for example: the log below would generate 77 points with the default `epsilon` value.

```
$ log2mission -start-offset 60 -end-offset -120 /t/inav-contrib/otxlogs/demolog.TXT
Flight   : MrPlane on 2021-04-08 13:24:07
Firmware : INAV 3.0.0 (fc0e5e274) MATEKF405 of Apr 7 2021 / 17:02:08
Size     : 19.36 MB
Log      : demolog.TXT / 1
Speed    : 28.0 m/s at 13:54
Range    : 17322 m at 14:22
Current  : 30.6 A at 00:05
Distance : 48437 m
Duration : 43:44
Altitude : 292.8 m at 25:42
Mission  : 56 points (reprocess: 1, epsilon: 0.018)
```

The output from this example would be `demolog.1.mission`

#### multirotor example

Using a old, contributed MR log, in quite a small area, with user specifed `epsilon`.

```
$ log2mission -epsilon 0.001 logfs.TXT
Log      : logfs.TXT / 1
Flight   :  on 2019-02-08 15:21:13
Firmware : INAV 2.1.0 (7bdd5967e) OMNIBUSF4V3 of Jan 22 2019 09:39:17
Size     : 32.03 MB
Current  : 23.5 A at 02:21
Distance : 1560 m
Duration : 04:10
Altitude : 52.5 m at 02:50
Speed    : 17.3 m/s at 02:38
Range    : 174 m at 01:22
Mission  : 13 points
```

13 points is a adequate mission to reproduce the flight.

Using an extreme user defined `epsilon` results in an excessive number of points:

```
$ log2mission -epsilon 0.00001 logfs.TXT
...
Mission  : 59 points (reprocess: 8, epsilon: 0.000105)
```

Whereas, with the default `epsilon` of 0.015, no useful mission is generated:

```
$ log2mission logfs.TXT
...
Mission  : 2 points
```
So some experimentation may be required to get a good mission, particularly for shorter MR flights. In particular, if reprocessing is indicated and the number of generated points is close to 60, then it's probably worth running again with a slightly larger `epsilon` than that shown in the output.

`log2mission` will make an attempt to resolve the "short/complex" 2 point results by increasing `epsilon` automatically.

## mission2kml

A standalone mission file to KML/Z converter is also provided.

```
$ mission2kml --help
Usage of mission2kml [options] mission_file
  -dms
    	Show positions as DMS (vice decimal degrees)
  -home string
    	Use home location

The home location is given as decimal degrees latitude and
longitude and optional altitude. The values should be separated by a single
separator, one of "/:; ,". If space is used, then the values must be enclosed
in quotes.

In locales where comma is used as decimal "point", then comma should not be
used as a separator.

If a syntactically valid home position is given, without altitude, an online
elevation service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789/104
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236,24
```

A KML file is generated to stdout, which may be redirected to a file, e.g:

```
$ mission2kml -home 54.125229,-4.730443 barrule-h.mission > mtest.kml
```

## Setting default options

Default settings may be set in a JSON formatted configuration file.

* On POSIX platforms (Linux, FreeBSD, MacOS), `$HOME/.config/fl2x/config.json`
* On Windows `%APPDIR%\fl2x\config.json`

The keys in the file are the relevant command line options, the following are recognised:

* `dms`
* `extrude`
* `kml`
* `rssi`
* `efficiency`
* `split-time`
* `home-alt`
* `blackbox-decode`
* `gradient`
* `outdir`
* `blt-vers`
* `type`
* `visibility`

For example:

```
{
    "dms": true,
    "extrude": true,
    "gradient": "yor",
    "efficiency": true
}
```

A warning will be displayed if the configuration file in not syntactically correct; in such cases its contents will be ignored. There is a [complete example](https://github.com/stronnag/bbl2kml/wiki/Sample-Config-file) in the wiki that can be used as a template.

Note also that the command interpreter allows `-flag` or `--flag` for any option.

## Limitations, Bugs, Bug Reporting

`flightlog2kml` aims to support as wide a range of inav firmware and log decoders as possible. During its development, inav has changed both the data logged and in some cases, the meaning of logged items; thus for versions of inav prior to 2.0, the reported flight mode might not be completely accurate. `flightlog2kml` is known to work with logs from 2015-10-30 (i.e. pre inav 1.0), and if you have a Blackbox log that is not decoded / visualised correctly, please raise a [Github issue](https://github.com/stronnag/bbl2kml/issues); this is a bug.

Due to the range of `inav` versions, `blackbox_decode` versions and supported operating systems, when reporting bugs, please include the following information in the Github issue:

* The version of `flightlog2kml` and `blackbox_decode`. Both applications have a `--help` option that should give the version numbers.
* The host operating system and version (e.g. "Debian Sid", "Windows 10", "MacOS 10.15").
* Provide the blackbox log that illustrates the problem. If you don't want to post the log into an essentially public forum (the Github issue), then please propose a private delivery channel.

## Building

Requires Go v1.13 or later.
Compiled with:

```
$ go build cmd/flightlog2kml/main.go
$ go build cmd/mission2kml/main.go
```

or more simply

```
make
```

**flightlog2kml** depends on [twpayne/go-kml](https://github.com/twpayne/go-kml), an outstanding open source Golang KML library.

`flightlog2kml` may be build for all OS for which a suitable Golang is available. It also requires inav's [blackbox_decode](https://github.com/iNavFlight/blackbox-tools); 0.4.5 (or future) is recommended; the minimum `blackbox_decode` version is 0.4.4. For Windows' users it is probably easiest to copy inav's `blackbox_decode.exe` into the same directory as `flightlog2kml.exe`.

Binaries are provided for common operating systems in the [Release folder](https://github.com/stronnag/bbl2kml/releases). Note that there is no binary for `fl2ltm`; this in "installed" manually as:

```
# cd <install location>
# ln -sf fl2mqtt fl2ltm
```

`fl2ltm` will be automatically detected by [mwp](https://github.com/stronnag/mwptools/) and used in preference to its older `replay_bbox_ltm.rb` and `otxlog` helpers.
