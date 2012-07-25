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
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
)

var (
	activeMqHost    = flag.String("activeMqHost", "localhost", "ActiveMQ host")
	activeMqPort    = flag.Int("activeMqPort", 8161, "ActiveMQ port")
	gangliaHost     = flag.String("gangliaHost", "localhost", "Ganglia host name/IP")
	gangliaPort     = flag.Int("gangliaPort", 8469, "Ganglia port")
	gangliaSpoof    = flag.String("gangliaSpoof", "", "Ganglia spoof string (IP:host)")
	gangliaGroup    = flag.String("gangliaGroup", "activemq", "Ganglia group name")
	gangliaInterval = flag.Int("gangliaInterval", 300, "Ganglia polling interval/metric TTL")
	verbose         = flag.Bool("verbose", false, "Verbose")
	ignoreQueues    = flag.String("ignoreQueues", "", "Substring to ignore in queue names")
)

func main() {
	flag.Parse()

	// Lookup host name
	gAddrs, err := net.LookupIP(*gangliaHost)
	if err != nil {
		panic(err.Error())
	}
	if len(gAddrs) < 1 {
		panic("Could not look up address " + *gangliaHost)
	}

	gm := gmetric.Gmetric{gAddrs[0], *gangliaPort, *gangliaHost, *gangliaSpoof}
	gm.SetVerbose(true)
	if *verbose {
		fmt.Printf("Established gmetric connection to %s\n", gAddrs[0])
	}

	q, err := GetQueues(*activeMqHost, *activeMqPort)
	if err != nil {
		panic(err)
	}
	for i := 0; i < len(q.Items); i++ {
		if *ignoreQueues == "" || !strings.Contains(q.Items[i].Name, *ignoreQueues) {
			fmt.Println(q.Items[i].Name)
			if *verbose {
        fmt.Printf("Sending queue_%s_size\n", q.Items[i].Name)
      }
			go gm.SendMetric(
				fmt.Sprintf("queue_%s_size", q.Items[i].Name),
				fmt.Sprint(q.Items[i].Stats.Size),
        gmetric.VALUE_UNSIGNED_INT, "size", gmetric.SLOPE_BOTH,
				uint32(*gangliaInterval), uint32(*gangliaInterval),
        *gangliaGroup)
			if *verbose {
        fmt.Printf("Sending queue_%s_consumers\n", q.Items[i].Name)
      }
			go gm.SendMetric(
				fmt.Sprintf("queue_%s_consumers", q.Items[i].Name),
				fmt.Sprint(q.Items[i].Stats.ConsumerCount),
        gmetric.VALUE_UNSIGNED_INT, "consumers", gmetric.SLOPE_BOTH,
				uint32(*gangliaInterval), uint32(*gangliaInterval),
        *gangliaGroup)
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
