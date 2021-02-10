package bltmqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"time"
	"strings"
	"strconv"
	"log"
	"io/ioutil"
	"math/rand"
	"os"
	"net/url"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	options "github.com/stronnag/bbl2kml/pkg/options"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

type MQTTClient struct {
	client mqtt.Client
	topic  string
}

func NewTlsConfig(cafile string) (*tls.Config, string) {
	if len(cafile) == 0 {
		return nil, "tcp"
	} else {
		certpool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cafile)
		if err != nil {
			log.Fatalln(err.Error())
		}
		certpool.AppendCertsFromPEM(ca)
		return &tls.Config{
			RootCAs:            certpool,
			InsecureSkipVerify: true, ClientAuth: tls.NoClientCert,
		},
			"ssl"
	}
}

func NewMQTTClient() *MQTTClient {
	var broker string
	var topic string
	var port int
	var cafile string
	var user string
	var passwd string

	rand.Seed(time.Now().UnixNano())

	u, err := url.Parse(options.Mqttopts)
	if err != nil {
		log.Fatal(err)
	}

	broker = u.Hostname()
	port, _ = strconv.Atoi(u.Port())

	if len(u.Path) > 0 {
		topic = u.Path[1:]
	}

	up := u.User
	user = up.Username()
	passwd, _ = up.Password()

	q := u.Query()
	ca := q["cafile"]
	if len(ca) > 0 {
		cafile = ca[0]
	}
	if broker == "" {
		broker = "broker.emqx.io"
	}

	if topic == "" {
		topic = fmt.Sprintf("org/mwptools/mqtt/loglayer/_%x", rand.Int())
		fmt.Fprintf(os.Stderr, "using random topic \"%s\"", topic)
	}

	if port == 0 {
		port = 1883
	}

	tlsconf, scheme := NewTlsConfig(cafile)
	if u.Scheme == "ws" {
		scheme = "ws"
	}
	if u.Scheme == "wss" {
		tlsconf = &tls.Config{RootCAs: nil, ClientAuth: tls.NoClientCert}
		scheme = "wss"
	}

	if tlsconf == nil && (u.Scheme == "mqtts" || u.Scheme == "ssl") {
		tlsconf = &tls.Config{RootCAs: nil, ClientAuth: tls.NoClientCert}
		scheme = "ssl"
	}

	if len(os.Getenv("NOVERIFYSSL")) > 0 && tlsconf != nil {
		tlsconf.InsecureSkipVerify = true
	}

	clientid := fmt.Sprintf("%x", rand.Int63())
	opts := mqtt.NewClientOptions()

	mpath := ""
	if scheme == "ws" || scheme == "wss" {
		mpath = "/mqtt"
	}
	hpath := fmt.Sprintf("%s://%s:%d%s", scheme, broker, port, mpath)

	opts.AddBroker(hpath)
	opts.SetTLSConfig(tlsconf)
	opts.SetClientID(clientid)
	opts.SetUsername(user)
	opts.SetPassword(passwd)
	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return &MQTTClient{client: client, topic: topic}
}

func (m *MQTTClient) publish(msg string) {
	token := m.client.Publish(m.topic, 0, false, msg)
	token.Wait()
}

/* Test brokers
   test.mosquitto.org 1883, 8883 8080, 8081 (ws)
   broker.hivemq.com  1883, 8000 (ws)
   broker.emqx.io    1883, 8883, 8083, 8084 (ws)
*/

func make_bullet_msg(b types.LogItem, homeamsl float64, elapsed int, ncells int) string {
	var sb strings.Builder

	sb.WriteString("flt:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')
	elapsed += 60
	sb.WriteString("ont:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')

	sb.WriteString("ran:")
	sb.WriteString(strconv.Itoa(int(b.Roll * 10)))
	sb.WriteByte(',')

	sb.WriteString("pan:")
	sb.WriteString(strconv.Itoa(int(b.Pitch * 10)))
	sb.WriteByte(',')

	sb.WriteString("hea:")
	sb.WriteString(strconv.Itoa(int(b.Cse)))
	sb.WriteByte(',')

	sb.WriteString("ggc:")
	sb.WriteString(strconv.Itoa(int(b.Cog)))
	sb.WriteByte(',')

	sb.WriteString("alt:")
	sb.WriteString(strconv.Itoa(int(b.Alt * 100)))
	sb.WriteByte(',')

	sb.WriteString("asl:")
	sb.WriteString(strconv.Itoa(int(b.GAlt)))
	sb.WriteByte(',')

	sb.WriteString("gsp:")
	sb.WriteString(strconv.Itoa(int(b.Spd * 100)))
	sb.WriteByte(',')

	sb.WriteString("bpv:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Volts * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", float64(b.Volts)))
	}
	sb.WriteByte(',')

	avc := b.Volts / float64(ncells)
	sb.WriteString("acv:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(avc * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", avc))
	}
	sb.WriteByte(',')

	sb.WriteString("cad:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Energy)))
	} else {
		sb.WriteString(fmt.Sprintf("%.0f", b.Energy))
	}
	sb.WriteByte(',')

	sb.WriteString("cud:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Amps * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", b.Amps))
	}
	sb.WriteByte(',')

	//	rssi := 100 * int(b.Rssi) / 255
	sb.WriteString("rsi:")
	sb.WriteString(strconv.Itoa(int(b.Rssi)))
	sb.WriteByte(',')

	sb.WriteString("gla:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Lat * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", b.Lat))
	}
	sb.WriteByte(',')

	sb.WriteString("glo:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Lon * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", b.Lon))
	}
	sb.WriteByte(',')

	sb.WriteString("gsc:")
	sb.WriteString(strconv.Itoa(int(b.Numsat)))
	sb.WriteByte(',')

	sb.WriteString("ghp:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Hdop)))
	} else {
		hdop := float64(b.Hdop) / 100.0
		sb.WriteString(fmt.Sprintf("%.1f", hdop))
	}
	sb.WriteByte(',')

	sb.WriteString("3df:")
	if b.Fix != 0 {
		sb.WriteString("1")
	} else {
		sb.WriteString("0")
	}
	sb.WriteByte(',')

	sb.WriteString("hds:")
	sb.WriteString(strconv.Itoa(int(b.Vrange)))
	sb.WriteByte(',')

	bearing := 180 - b.Bearing
	if bearing < 0 {
		bearing += 360
	}

	sb.WriteString("hdr:")
	sb.WriteString(strconv.Itoa(int(bearing)))
	sb.WriteByte(',')

	sb.WriteString("trp:")
	sb.WriteString(strconv.Itoa(b.Throttle))
	sb.WriteByte(',')

	fs := (b.Status & 2) >> 1
	sb.WriteString("fs:")
	sb.WriteString(strconv.Itoa(int(fs)))
	sb.WriteByte(',')

	armed := b.Status & 1
	sb.WriteString(fmt.Sprintf("arm:%d", armed))
	return sb.String()
}

func make_bullet_home(hlat float64, hlon float64, halt float64) string {
	var sb strings.Builder
	sb.WriteString("cs:JRandomUAV,")
	sb.WriteString("hla:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(hlat * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", hlat))
	}
	sb.WriteByte(',')
	sb.WriteString("hlo:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(hlon * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", hlon))
	}
	sb.WriteByte(',')
	sb.WriteString("hal:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(halt * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.0f", halt))
	}
	return sb.String()
}

func make_bullet_mode(mode string, ncells int, hwfail bool) string {
	var sb strings.Builder
	if ncells > 0 {
		sb.WriteString("bcc:")
		sb.WriteString(strconv.Itoa(ncells))
		sb.WriteByte(',')
	}

	sb.WriteString("ftm:")
	sb.WriteString(mode)
	hwok := 1
	if hwfail {
		hwok = 0
	}
	sb.WriteString(fmt.Sprintf(",css:3,hwh:%d,", hwok))
	return sb.String()
}

func get_cells(vbat float64) int {
	ncell := 0
	for i := 1; i < 10; i++ {
		vmin := 3.0 * float64(i)
		vmax := 4.22 * float64(i)
		if vbat < vmax && vbat > vmin {
			ncell = i
			break
		}
	}
	return ncell
}

func MQTTGen(s types.LogSegment) {
	ncells := 0
	c := NewMQTTClient()
	var lastm time.Time
	laststat := uint8(255)
	fmode := ""
	mstrs := []string{}
	wps := ""

	if len(options.Mission) > 0 {
		_, ms, err := mission.Read_Mission_File(options.Mission)
		if err == nil {
			var sb strings.Builder
			for k, mi := range ms.MissionItems {
				if geo.Getfrobnication() && mi.Is_GeoPoint() {
					ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, _ = geo.Frobnicate_move(ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, 0)
				}
				act, ok := mission.ActionMap[mi.Action]
				if !ok {
					act = 1
				}
				sb.Reset()
				sb.WriteString(fmt.Sprintf("wpno:%d,la:%d,lo:%d,al:%d,ac:%d,", mi.No,
					int(10000000*ms.MissionItems[k].Lat), int(10000000*ms.MissionItems[k].Lon),
					mi.Alt*100, act))
				if mi.P1 != 0 {
					sb.WriteString(fmt.Sprintf("p1:%d,", mi.P1))
				}
				if mi.P2 != 0 {
					sb.WriteString(fmt.Sprintf("p2:%d,", mi.P2))
				}
				if mi.P3 != 0 {
					sb.WriteString(fmt.Sprintf("p3:%d,", mi.P3))
				}
				if k == len(ms.MissionItems)-1 {
					sb.WriteString("f:165")
				}
				mstrs = append(mstrs, sb.String())
			}
			wps = fmt.Sprintf("wpc:%d,wpv:1,", len(ms.MissionItems))
		} else {
			fmt.Fprintf(os.Stderr, "* Failed to read mission file %s\n", options.Mission)
		}
	}

	c.publish("wpc:0,wpv:0,flt:0,ont:0")
	st := time.Now()
	for i, b := range s.L.Items {
		now := time.Now()
		et := int(now.Sub(st).Seconds())
		stat := b.Status >> 2

		if ncells == 0 {
			ncells = get_cells(b.Volts)
		}

		if b.Fmode != laststat {
			if options.Bulletvers == 2 {
				switch b.Fmode {
				case types.FM_MANUAL:
					fmode = "1"
				case types.FM_ANGLE:
					fmode = "9"
				case types.FM_HORIZON:
					fmode = "10"
				case types.FM_ACRO:
					fmode = "11"
				case types.FM_AH:
					fmode = "8"
				case types.FM_PH:
					fmode = "4"
				case types.FM_WP:
					fmode = "7"
				case types.FM_RTH:
					fmode = "2"
				case types.FM_CRUISE3D:
					fmode = "5"
				case types.FM_LAUNCH:
					fmode = "9"
				default:
					fmode = "11"
				}
			} else {
				switch b.Fmode {
				case types.FM_MANUAL:
					fmode = "MANU"
				case types.FM_ANGLE:
					fmode = "ANGL"
				case types.FM_HORIZON:
					fmode = "HOR"
				case types.FM_ACRO:
					fmode = "ACRO"
				case types.FM_AH:
					fmode = "A H"
				case types.FM_PH:
					fmode = "P H"
				case types.FM_WP:
					fmode = "WP"
				case types.FM_RTH:
					fmode = "RTH"
				case types.FM_CRUISE3D:
					fmode = "3CRS"
				case types.FM_LAUNCH:
					fmode = "LNCH"
				default:
					fmode = "ACRO"
				}
				if stat != 0 {
					fmode = "!FS!"
				}
			}
			laststat = b.Fmode
			msg := make_bullet_mode(fmode, ncells, b.HWfail)
			c.publish(msg)
		}

		if i%10 == 0 {
			msg := make_bullet_mode(fmode, ncells, b.HWfail)
			c.publish(msg)
			msg = make_bullet_home(s.H.HomeLat, s.H.HomeLon, s.H.HomeAlt)
			c.publish(msg)
			if len(mstrs) > 0 && i%20 == 0 {
				for _, str := range mstrs {
					c.publish(str)
				}
				c.publish(wps)
			}
		}

		msg := make_bullet_msg(b, s.H.HomeAlt, et, ncells)
		c.publish(msg)
		if !lastm.IsZero() {
			tdiff := b.Utc.Sub(lastm)
			time.Sleep(tdiff)
		}
		lastm = b.Utc
	}
}
