// GANGLIA-ACTIVEMQ
// https://github.com/jbuchbinder/ganglia-activemq

package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/jbuchbinder/go-gmetric/gmetric"
	"io/ioutil"
	"log/syslog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

var (
	activeMqHost    = flag.String("activeMqHost", "localhost", "ActiveMQ host")
	activeMqPort    = flag.Int("activeMqPort", 8161, "ActiveMQ port")
	gangliaHost     = flag.String("gangliaHost", "localhost", "Ganglia host name/IP, can be multiple comma separated")
	gangliaPort     = flag.Int("gangliaPort", 8649, "Ganglia port")
	gangliaSpoof    = flag.String("gangliaSpoof", "", "Ganglia spoof string (IP:host)")
	gangliaGroup    = flag.String("gangliaGroup", "activemq", "Ganglia group name")
	gangliaInterval = flag.Int("gangliaInterval", 300, "Ganglia polling interval/metric TTL")
	vdedServer      = flag.String("vdedServer", "", "VDED server (default not used)")
	verbose         = flag.Bool("verbose", false, "Verbose")
	ignoreQueues    = flag.String("ignoreQueues", "", "Substring to ignore in queue names")
	gm              gmetric.Gmetric
	log, _          = syslog.New(syslog.LOG_DEBUG, "ganglia-activemq")
)

func main() {
	flag.Parse()

	gm = gmetric.Gmetric{
		Spoof: *gangliaSpoof,
		Host:  *gangliaSpoof,
	}
	gm.SetLogger(log)
	if *verbose {
		gm.SetVerbose(true)
	} else {
		gm.SetVerbose(false)
	}

	if strings.Contains(*gangliaHost, ",") {
		segs := strings.Split(*gangliaHost, ",")
		for i := 0; i < len(segs); i++ {
			gIP, err := net.ResolveIPAddr("ip4", segs[i])
			if err != nil {
				panic(err.Error())
			}
			gm.AddServer(gmetric.GmetricServer{
				Server: gIP.IP,
				Port:   *gangliaPort,
			})
		}
	} else {
		// Lookup host name
		gIP, err := net.ResolveIPAddr("ip4", *gangliaHost)
		if err != nil {
			panic(err.Error())
		}
		gm.AddServer(gmetric.GmetricServer{
			Server: gIP.IP,
			Port:   *gangliaPort,
		})
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	q, err := getQueues(*activeMqHost, *activeMqPort)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(q.Items); i++ {
		if *ignoreQueues == "" || !strings.Contains(q.Items[i].Name, *ignoreQueues) {
			sName := strings.Replace(q.Items[i].Name, ".", "_", -1)
			log.Debug("Processing queue " + q.Items[i].Name)
			if *verbose {
				fmt.Printf("Sending queue_%s_size\n", q.Items[i].Name)
			}
			gm.SendMetric(
				fmt.Sprintf("queue_%s_size", sName),
				fmt.Sprint(q.Items[i].Stats.Size),
				gmetric.VALUE_UNSIGNED_INT, "size", gmetric.SLOPE_BOTH,
				uint32(*gangliaInterval), uint32(*gangliaInterval)*2,
				*gangliaGroup)
			if *verbose {
				fmt.Printf("Sending queue_%s_consumers\n", q.Items[i].Name)
			}
			gm.SendMetric(
				fmt.Sprintf("queue_%s_consumers", sName),
				fmt.Sprint(q.Items[i].Stats.ConsumerCount),
				gmetric.VALUE_UNSIGNED_INT, "consumers", gmetric.SLOPE_BOTH,
				uint32(*gangliaInterval), uint32(*gangliaInterval)*2,
				*gangliaGroup)
			if *vdedServer != "" {
				if *verbose {
					fmt.Printf("VDED submitting %s queue_%s_enqueue : %d\n", hostname, sName, q.Items[i].Stats.EnqueueCount)
				}
				go submitVded(hostname, fmt.Sprintf("queue_%s_enqueue", sName), q.Items[i].Stats.EnqueueCount)
				if *verbose {
					fmt.Printf("VDED submitting %s queue_%s_dequeue : %d\n", hostname, sName, q.Items[i].Stats.DequeueCount)
				}
				go submitVded(hostname, fmt.Sprintf("queue_%s_dequeue", sName), q.Items[i].Stats.DequeueCount)
			}
		}
	}
}

func getQueues(host string, port int) (q queues, e error) {
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
	var obj queues
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

func submitVded(host string, metricName string, value int64) (e error) {
	if *vdedServer == "" {
		return errors.New("No VDED server, cannot continue")
	}

	url := fmt.Sprintf("http://%s:48333/vector?host=%s&vector=delta_%s&value=%d&ts=%d&submit_metric=1", *vdedServer, host, metricName, value, time.Now().Unix())
	log.Debug("VDED: " + url)

	c := http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(time.Duration(5) * time.Second),
		},
	}

	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debug("VDED returned: " + string(body))

	return
}

func timeoutDialer(ns time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		c, err := net.Dial(netw, addr)
		if err != nil {
			return nil, err
		}
		c.SetDeadline(time.Now().Add(ns))
		return c, nil
	}
}
