package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/mmcquillan/joda"
	"github.com/mmcquillan/matcher"
	"github.com/tkuchiki/go-timezone"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
	"gopkg.in/yaml.v2"
)

func main() {

	// list function
	match, _, _ := matcher.Matcher("<bin> list ", strings.Join(os.Args, " "))
	if match {
		tz := timezone.GetAllTimezones()
		for k, v := range tz {
			fmt.Printf("%s\n", k)
			for _, t := range v {
				fmt.Printf(" - %s\n", t)
			}
		}
		os.Exit(0)
	}

	// init zones
	var zones []Zone

	// capture input
	match, _, values := matcher.Matcher("<bin> [zones] [--date] [--24]", strings.Join(os.Args, " "))
	if !match || values["zones"] == "help" {
		fmt.Println("tz list")
		fmt.Println("tz [zones] [--date] [--24]")
		fmt.Println("")
		fmt.Println("examples:")
		fmt.Println("  tz UTC")
		fmt.Println("  tz UTC,Local")
		fmt.Println("  tz America/New_York")
		fmt.Println("  tz America/New_York:Matt")
		fmt.Println("  tz America/New_York:Matt:8:17")
		fmt.Println("  tz UTC,America/New_York:Matt:8:17")
		os.Exit(0)
	}

	// set 24
	block := 8
	format := "hha"
	if os.Getenv("TZ_24") == "true" || values["24"] == "true" {
		block = 6
		format = "H"
	}

	// set date
	date := false
	if os.Getenv("TZ_DATE") == "true" || values["date"] == "true" {
		date = true
	}

	// users file
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println("ERROR: No home dir - " + err.Error())
	}
	file := path.Join(home, ".tz")

	// set zones
	if values["zones"] != "" {
		for _, tz := range strings.Split(values["zones"], ",") {
			zones = append(zones, splitInput(tz))
		}
	} else if _, err := os.Stat(file); err == nil {
		tzFile, err := ioutil.ReadFile(file)
		err = yaml.Unmarshal([]byte(tzFile), &zones)
		if err != nil {
			fmt.Println("ERROR: Unmarshal error - " + err.Error())
		}
	} else if os.Getenv("TZ_ZONES") != "" {
		for _, tz := range strings.Split(os.Getenv("TZ_ZONES"), ",") {
			zones = append(zones, splitInput(tz))
		}
	} else {
		zones = append(zones, splitInput("UTC"))
		zones = append(zones, splitInput("Local"))
	}

	// max name
	name := 0
	for i, z := range zones {
		if z.Name == "" {
			zones[i].Name = z.TZ
		}
		if len(z.Name) > name {
			name = len(z.Name)
		}
	}

	// spacing
	width, _ := terminal.Width()
	remWidth := int(width) - (name + 2)
	full := (remWidth - (remWidth % block)) / block
	half := (full - (full % 2)) / 2

	// colors
	inactive := color.New(color.BgWhite).Add(color.FgBlack)
	active := color.New(color.BgGreen)
	now := color.New(color.BgBlue)
	nope := color.New(color.BgRed)
	info := color.New(color.FgRed)

	// set time
	n := time.Now().UTC()

	// output
	fmt.Printf("\n")
	for _, z := range zones {
		offset, match := findOffset(z.TZ)
		if match {
			if z.Highlight {
				info.Printf(" %s%s ", strings.Repeat(" ", name-len(z.Name)), z.Name)
			} else {
				fmt.Printf(" %s%s ", strings.Repeat(" ", name-len(z.Name)), z.Name)
			}
			for i := -half + 1; i <= half; i++ {
				t := n.Add(time.Second * time.Duration((i*3600)+offset))
				if i == 0 {
					if date {
						now.Printf(" %s - %s ", t.Format(joda.Format("MM/dd")), t.Format(joda.Format(format)))
					} else {
						now.Printf(" %s ", t.Format(joda.Format(format)))
					}
				} else if t.Hour() >= z.Start && t.Hour() <= z.End {
					active.Printf(" %s ", t.Format(joda.Format(format)))
				} else {
					inactive.Printf(" %s ", t.Format(joda.Format(format)))
				}
				fmt.Printf(" ")
			}
		} else {
			fmt.Printf(" %s%s ", strings.Repeat(" ", name-len(z.Name)), z.Name)
			nope.Printf(" Cannot find timezone: %s ", z.TZ)
		}
		fmt.Printf("\n\n")
	}

}

func splitInput(input string) (z Zone) {

	// cleanup
	input = strings.TrimSpace(input)

	// default
	z = Zone{
		TZ:        input,
		Name:      input,
		Start:     25,
		End:       25,
		Highlight: false,
	}

	// do we split
	if strings.Contains(input, ":") {
		p := strings.Split(input, ":")
		switch len(p) {
		case 2:
			z.TZ = strings.TrimSpace(p[0])
			z.Name = strings.TrimSpace(p[1])
		case 3:
			z.TZ = strings.TrimSpace(p[0])
			z.Name = strings.TrimSpace(p[1])
			if val, err := strconv.ParseInt(p[2], 10, 32); err == nil {
				z.Start = int(val)
			}
		case 4:
			z.TZ = strings.TrimSpace(p[0])
			z.Name = strings.TrimSpace(p[1])
			if val, err := strconv.ParseInt(p[2], 10, 32); err == nil {
				z.Start = int(val)
			}
			if val, err := strconv.ParseInt(p[3], 10, 32); err == nil {
				z.End = int(val)
			}
		}
	}

	if strings.HasPrefix(z.Name, "@") {
		z.Name = strings.Replace(z.Name, "@", "", -1)
		z.Highlight = true
	}

	// return
	return z

}

func findOffset(tz string) (offset int, match bool) {

	// init
	offset = 0
	match = false

	// shorthand timezones
	tzShorthand := strings.ToUpper(tz)
	switch tzShorthand {
	case "LOCAL":
		tz = "Local"
	case "EASTERN":
		tz = "America/New_York"
	case "CENTRAL":
		tz = "America/Chicago"
	case "MOUNTAIN":
		tz = "America/Denver"
	case "PACIFIC":
		tz = "America/Los_Angeles"
	}

	// first, std timezone lib
	loc, err := time.LoadLocation(tz)
	if err == nil {
		match = true
		n := time.Now().UTC().In(loc)
		_, offset = n.Zone()
	}

	// second, tz lib
	if !match {
		offset, err = timezone.GetOffset(tz)
		if err == nil {
			match = true
		}
	}

	// return
	return offset, match
}

type Zone struct {
	TZ        string `yaml:"tz"`
	Name      string `yaml:"name"`
	Start     int    `yaml:"start"`
	End       int    `yaml:"end"`
	Highlight bool   `yaml:"highlight"`
}
