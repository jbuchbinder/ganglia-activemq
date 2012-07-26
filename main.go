// GANGLIA-ACTIVEMQ
// https://github.com/jbuchbinder/ganglia-activemq

package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/jbuchbinder/go-gmetric/gmetric"
	"io/ioutil"
	"log/syslog"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

var (
	activeMqHost    = flag.String("activeMqHost", "localhost", "ActiveMQ host")
	activeMqPort    = flag.Int("activeMqPort", 8161, "ActiveMQ port")
	gangliaHost     = flag.String("gangliaHost", "localhost", "Ganglia host name/IP, can be multiple comma separated")
	gangliaPort     = flag.Int("gangliaPort", 8649, "Ganglia port")
	gangliaSpoof    = flag.String("gangliaSpoof", "", "Ganglia spoof string (IP:host)")
	gangliaGroup    = flag.String("gangliaGroup", "activemq", "Ganglia group name")
	gangliaInterval = flag.Int("gangliaInterval", 300, "Ganglia polling interval/metric TTL")
	verbose         = flag.Bool("verbose", false, "Verbose")
	ignoreQueues    = flag.String("ignoreQueues", "", "Substring to ignore in queue names")
	gm              []gmetric.Gmetric
	log, _          = syslog.New(syslog.LOG_DEBUG, "ganglia-activemq")
)

func main() {
	flag.Parse()

	var gIPs []net.IPAddr

	if strings.Contains(*gangliaHost, ",") {
		gIPs = make([]net.IPAddr, strings.Count(*gangliaHost, ",")+1)
		gm = make([]gmetric.Gmetric, strings.Count(*gangliaHost, ",")+1)
		segs := strings.Split(*gangliaHost, ",")
		for i := 0; i < len(segs); i++ {
			gIP, err := net.ResolveIPAddr("ip4", segs[i])
			if err != nil {
				panic(err.Error())
			}
			gIPs[i] = *gIP
		}
	} else {
		gIPs = make([]net.IPAddr, 1)
		gm = make([]gmetric.Gmetric, 1)
		// Lookup host name
		gIP, err := net.ResolveIPAddr("ip4", *gangliaHost)
		if err != nil {
			panic(err.Error())
		}
		gIPs[0] = *gIP
	}

	fmt.Printf("len(gIPs) = %d\n", len(gIPs))
	for i := 0; i < len(gIPs); i++ {
		gm[i] = gmetric.Gmetric{gIPs[i].IP, *gangliaPort, *gangliaSpoof, *gangliaSpoof}
		gm[i].SetLogger(log)
		if *verbose {
			gm[i].SetVerbose(true)
			fmt.Printf("Established gmetric connection to %s:%d\n", gIPs[i].IP, *gangliaPort)
		} else {
			gm[i].SetVerbose(false)
		}
	}

	q, err := GetQueues(*activeMqHost, *activeMqPort)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(q.Items); i++ {
		if *ignoreQueues == "" || !strings.Contains(q.Items[i].Name, *ignoreQueues) {
			sName := strings.Replace(q.Items[i].Name, ".", "_", -1)
			log.Debug("Processing queue " + q.Items[i].Name)
			for j := 0; j < len(gm); j++ {
				if *verbose {
					fmt.Printf("Sending queue_%s_size to %s\n", q.Items[i].Name, gIPs[j].IP)
				}
				gm[j].SendMetric(
					fmt.Sprintf("queue_%s_size", sName),
					fmt.Sprint(q.Items[i].Stats.Size),
					gmetric.VALUE_UNSIGNED_INT, "size", gmetric.SLOPE_BOTH,
					uint32(*gangliaInterval), uint32(*gangliaInterval)*2,
					*gangliaGroup)
				if *verbose {
					fmt.Printf("Sending queue_%s_consumers to %s\n", q.Items[i].Name, gIPs[j].IP)
				}
				gm[j].SendMetric(
					fmt.Sprintf("queue_%s_consumers", sName),
					fmt.Sprint(q.Items[i].Stats.ConsumerCount),
					gmetric.VALUE_UNSIGNED_INT, "consumers", gmetric.SLOPE_BOTH,
					uint32(*gangliaInterval), uint32(*gangliaInterval)*2,
					*gangliaGroup)
			}
		}
	}
}

func GetQueues(host string, port int) (q Queues, e error) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s:%d/admin/xml/queues.jsp", host, port)
	if *verbose {
		fmt.Println("url = '" + url + "'")
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		e = err
		return
	}
	if *verbose {
		dump, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(dump))
	}
	req.Header.Set("User-Agent", "ganglia-activemq")

	res, err := client.Do(req)
	if err != nil {
		e = err
		return
	}
	defer res.Body.Close()

	// Extract user resource from body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		e = err
		return
	}
	var obj Queues
	err = xml.Unmarshal(body, &obj)
	if err != nil {
		e = err
		return
	}

	if *verbose {
		m, err := json.MarshalIndent(obj, " ", "  ")
		if err != nil {
			e = err
			return
		}
		fmt.Println(string(m))
	}

	q = obj
	return
}
