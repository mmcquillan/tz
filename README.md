# tz

This is a simple timezone visualizer.

```
tz list
tz [zones] [--24]

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

You can also set via envrionment variables:


`TZ_ZONES="UTC,America/New_York:Matt:8:17,America/Los_Angeles:Jack:8:17"`


`TZ_24=true`


