package sitlgen

import (
	"bufio"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SimMeta struct {
	sitl     string
	ip       string
	port     string
	path     string
	eeprom   string
	mintime  int
	failmode uint16
}

func read_cfg() SimMeta {
	sitl := SimMeta{}
	cdir := types.GetConfigDir()
	fn := filepath.Join(cdir, "fl2sitl.conf")
	r, err := os.Open(fn)
	if err == nil {
		defer r.Close()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			l := scanner.Text()
			l = strings.TrimSpace(l)
			if !(len(l) == 0 || strings.HasPrefix(l, "#") || strings.HasPrefix(l, ";")) {
				parts := strings.SplitN(l, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					switch key {
					case "sitl":
						sitl.sitl = val
					case "simip":
						sitl.ip = val
					case "simport":
						sitl.port = val
					case "eeprom-path":
						sitl.path = val
					case "default-eeprom":
						sitl.eeprom = val
					case "min-time":
						sitl.mintime, _ = strconv.Atoi(val)
					case "failmode":
						if val[0] == 'i' {
							sitl.failmode = 0
						} else if val[0] == 'n' {
							sitl.failmode = 0xd0d0
						} else {
							tmp, _ := strconv.Atoi(val)
							sitl.failmode = uint16(tmp)
						}
					}
				}
			}
		}
	} else {
		log.Fatal("%s : %v\n", fn, err)
	}
	return sitl
}
