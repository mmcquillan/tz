# tz - A simple CLI timezone visualizer.

#### Installation

```bash
go get "github.com/mmcquillan/tz"
```

#### Usage
```
tz list
tz [zones] [--date] [--24]

examples:
  tz UTC
  tz UTC,Local
  tz America/New_York
  tz America/New_York:Matt
  tz America/New_York:Matt:8:17
  tz UTC,America/New_York:Matt:8:17
```

Zone format is a comma delimited set of:

```
<timezone>[:name][:start time][:end time]
```

You can set zones in your home director `~/.tz`:

```
- tz: "UTC"
- tz: "local"
  name: "Matt"
  start: 8
  end: 17
- tz: "America/Los_Angeles"
  name: "Jack"
  start: 8
  end: 17
```


You can also set via envrionment variables:


`TZ_ZONES="UTC,America/New_York:Matt:8:17,America/Los_Angeles:Jack:8:17"`


`TZ_24=true`


`TZ_DATE=true`


![example](https://raw.githubusercontent.com/mmcquillan/tz/master/example.png)

