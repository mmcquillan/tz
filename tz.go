package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mmcquillan/joda"
	"github.com/mmcquillan/matcher"
	"github.com/tkuchiki/go-timezone"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
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
	var zones []zone

	// capture input
	match, _, values := matcher.Matcher("<bin> [zones] [--24]", strings.Join(os.Args, " "))
	if !match || values["zones"] == "help" {
		fmt.Println("tz list")
		fmt.Println("tz [zones] [--24]")
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
	block := 7
	format := "hha"
	if os.Getenv("TZ_24") == "true" || values["24"] == "true" {
		block = 5
		format = "H"
	}

	// set zones
	if os.Getenv("TZ_ZONES") != "" {
		for _, tz := range strings.Split(os.Getenv("TZ_ZONES"), ",") {
			zones = append(zones, splitInput(tz))
		}
	} else if values["zones"] != "" {
		for _, tz := range strings.Split(values["zones"], ",") {
			zones = append(zones, splitInput(tz))
		}
	} else {
		zones = append(zones, splitInput("UTC"))
		zones = append(zones, splitInput("Local"))
	}

	// max name
	name := 0
	for _, z := range zones {
		if len(z.name) > name {
			name = len(z.name)
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

	// set time
	n := time.Now().UTC()

	// output
	fmt.Printf("\n")
	for _, z := range zones {
		offset, match := findOffset(z.tz)
		if match {
			fmt.Printf(" %s%s ", strings.Repeat(" ", name-len(z.name)), z.name)
			for i := -half + 1; i <= half; i++ {
				t := n.Add(time.Second * time.Duration((i*3600)+offset))
				if i == 0 {
					now.Printf(" %s ", t.Format(joda.Format(format)))
				} else if t.Hour() >= z.start && t.Hour() <= z.end {
					active.Printf(" %s ", t.Format(joda.Format(format)))
				} else {
					inactive.Printf(" %s ", t.Format(joda.Format(format)))
				}
				fmt.Printf(" ")
			}
		} else {
			fmt.Printf(" %s%s ", strings.Repeat(" ", name-len(z.name)), z.name)
			nope.Printf(" Cannot find timezone: %s ", z.tz)
		}
		fmt.Printf("\n\n")
	}

}

func splitInput(input string) (z zone) {

	// cleanup
	input = strings.TrimSpace(input)

	// default
	z = zone{
		tz:    input,
		name:  input,
		start: 25,
		end:   25,
	}

	// do we split
	if strings.Contains(input, ":") {
		p := strings.Split(input, ":")
		switch len(p) {
		case 2:
			z.tz = strings.TrimSpace(p[0])
			z.name = strings.TrimSpace(p[1])
		case 3:
			z.tz = strings.TrimSpace(p[0])
			z.name = strings.TrimSpace(p[1])
			if val, err := strconv.ParseInt(p[2], 10, 32); err == nil {
				z.start = int(val)
			}
		case 4:
			z.tz = strings.TrimSpace(p[0])
			z.name = strings.TrimSpace(p[1])
			if val, err := strconv.ParseInt(p[2], 10, 32); err == nil {
				z.start = int(val)
			}
			if val, err := strconv.ParseInt(p[3], 10, 32); err == nil {
				z.end = int(val)
			}
		}
	}

	// return
	return z

}

func findOffset(tz string) (offset int, match bool) {

	// init
	offset = 0
	match = false

	// cleanup
	if strings.ToUpper(tz) == "LOCAL" {
		tz = "Local"
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

type zone struct {
	tz    string
	name  string
	start int
	end   int
}
