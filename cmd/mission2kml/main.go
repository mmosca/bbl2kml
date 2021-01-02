package main

import (
	"flag"
	"os"
	"log"
	"fmt"
	"strings"
	"strconv"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
)

var (
	dms     bool
	homepos string
)

func split(s string, separators []rune) []string {
	f := func(r rune) bool {
		for _, s := range separators {
			if r == s {
				return true
			}
		}
		return false
	}
	return strings.FieldsFunc(s, f)
}

func main() {
	hlat := 0.0
	hlon := 0.0
	use_elev := false

	flag.Usage = func() {
		extra := `The home location is given as decimal degrees latitude and
longitude. The values should be separated by a single separator, one
of "/:; ,". If space is used, then the values must be enclosed in
quotes. In locales where comma is used as decimal "point", then it
should not be used as a separator.

If a syntactically valid home postion is given, an online elevation
service is used to adjust mission elevations in the KML.

Examples:
    -home 54.353974/-4.5236
    --home 48,9975:2,5789
    -home 54.353974;-4.5236
    --home "48,9975 2,5789"
    -home 54.353974,-4.5236
`
		fmt.Fprintf(os.Stderr, "Usage of missionkml [options] mission_file\n")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, extra)
	}

	defs := os.Getenv("BBL2KML_OPTS")
	dms = strings.Contains(defs, "-dms")

	flag.BoolVar(&dms, "dms", dms, "Show positions as DMS (vice decimal degrees)")
	flag.StringVar(&homepos, "home", homepos, "Use home location")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		flag.Usage()
		os.Exit(-1)
	}

	if len(homepos) > 0 {
		parts := split(homepos, []rune{'/', ':', ';', ' ', ','})
		if len(parts) == 2 {
			var err error
			hlat, err = strconv.ParseFloat(parts[0], 64)
			if err == nil {
				hlon, err = strconv.ParseFloat(parts[1], 64)
				if (hlat != 0.0 && hlon != 0.0) && hlat <= 90.0 && hlat >= -90 &&
					hlon <= 180.0 && hlat >= -180 {
					use_elev = true
				}
			}
		}
	}

	_, m, err := mission.Read_Mission_File(files[0])
	if m != nil && err == nil {
		m.Dump(dms, use_elev, hlat, hlon)
	}
	if err != nil {
		log.Fatal(err)
	}
}
