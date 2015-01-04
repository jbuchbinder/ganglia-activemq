# GANGLIA-ACTIVEMQ

[![Gobuild Download](http://gobuild.io/badge/github.com/jbuchbinder/ganglia-activemq/downloads.svg)](http://gobuild.io/github.com/jbuchbinder/ganglia-activemq)

 * Github: https://github.com/jbuchbinder/ganglia-activemq
 * Twitter: [@jbuchbinder](https://twitter.com/jbuchbinder)

Allows ActiveMQ monitoring stats to be pulled into Ganglia. It can optionally
use [VDED](https://github.com/jbuchbinder/vded) to aggregate certain counters.

## BUILDING

```
go get github.com/jbuchbinder/go-gmetric/gmetric
go install github.com/jbuchbinder/go-gmetric/gmetric
go build
```

## USAGE

```
Usage of ganglia-activemq:
  -activeMqHost="localhost": ActiveMQ host
  -activeMqPort=8161: ActiveMQ port
  -gangliaGroup="activemq": Ganglia group name
  -gangliaHost="localhost": Ganglia host name/IP, can be multiple comma separated
  -gangliaInterval=300: Ganglia polling interval/metric TTL
  -gangliaPort=8649: Ganglia port
  -gangliaSpoof="": Ganglia spoof string (IP:host)
  -ignoreQueues="": Substring to ignore in queue names
  -vdedServer="": VDED server (default not used)
  -verbose=false: Verbose
```

