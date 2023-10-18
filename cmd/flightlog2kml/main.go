package main

import (
	"fmt"
	"github.com/yookoala/realpath"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

import (
	"aplog"
	"bbl"
	"bltlog"
	"geo"
	"kmlgen"
	"options"
	"otx"
	"types"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s commit:%s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}

func main() {
	dump_log := os.Getenv("DUMP_LOG") != ""
	files, _ := options.ParseCLI(getVersion)
	if len(files) == 0 {
		if len(options.Config.Mission) > 0 {
			outms := kmlgen.GenKmlName(options.Config.Mission, options.Config.MissionIndex)
			kmlgen.GenerateMissionOnly(outms)
			show_output(outms)
		} else {
			options.Usage()
		}
		os.Exit(1)
	}

	geo.Frobnicate_init()

	var lfr types.FlightLog
	for _, fn := range files {
		ftype := types.EvinceFileType(fn)
		switch ftype {
		case types.IS_OTX:
			l := otx.NewOTXReader(fn)
			lfr = &l
		case types.IS_BBL:
			l := bbl.NewBBLReader(fn)
			lfr = &l
		case types.IS_BLT:
			l := bltlog.NewBLTReader(fn)
			lfr = &l
		case types.IS_AP:
			l := aplog.NewAPReader(fn)
			lfr = &l
		default:
			log.Fatalf("%s: unknown log format\n", fn)
		}

		metas, err := lfr.GetMetas()
		if err == nil {
			if options.Config.Dump {
				lfr.Dump()
				os.Exit(0)
			}
			options.Config.Tmpdir, err = ioutil.TempDir("", ".fl2x")
			if err != nil {
				log.Fatalf("fl2x: %+v\n", err)
			}
			defer os.RemoveAll(options.Config.Tmpdir)

			for _, b := range metas {
				outfn := ""
				if (options.Config.Idx == 0 || options.Config.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
					for k, v := range b.Summary() {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					ls, res := lfr.Reader(b, nil)
					if res {
						if dump_log {
							for _, b := range ls.L.Items {
								fmt.Fprintf(os.Stderr, "%+v\n", b)
							}
						} else if options.Config.Summary == false {
							outfn = kmlgen.GenKmlName(b.Logname, b.Index)
							kmlgen.GenerateKML(ls.H, ls.L, outfn, b, ls.M)
						}
					}
					for k, v := range ls.M {
						fmt.Printf("%-8.8s : %s\n", k, v)
					}
					if s, ok := b.ShowDisarm(); ok {
						fmt.Printf("%-8.8s : %s\n", "Disarm", s)
					}
					if !res {
						fmt.Fprintf(os.Stderr, "*** skipping KML/Z for log  with no valid geospatial data\n")
					} else {
						show_output(outfn)
					}
					fmt.Println()
				}
			}
		} else {
			log.Fatalf("fl2x: %+v\n", err)
		}
	}
}

func show_output(outfn string) {
	if outfn != "" {
		rp, err := realpath.Realpath(outfn)
		if err != nil || rp == "" {
			fmt.Printf("%-8.8s : <%s> <%s>\n", "RealPath", rp, err)
			rp = outfn
		}
		fmt.Printf("%-8.8s : %s\n", "Output", rp)
	}
}
