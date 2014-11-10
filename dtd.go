// GANGLIA-ACTIVEMQ
// https://github.com/jbuchbinder/ganglia-activemq

package main

type queues struct {
	Items []queue `xml:"queue" json:"queue"`
}

type queue struct {
	Name  string     `xml:"name,attr" json:"name"`
	Stats queueStats `xml:"stats" json:"stats"`
}

type queueStats struct {
	Size          int64 `xml:"size,attr" json:"size"`
	ConsumerCount int64 `xml:"consumerCount,attr" json:"consumerCount"`
	EnqueueCount  int64 `xml:"enqueueCount,attr" json:"enqueueCount"`
	DequeueCount  int64 `xml:"dequeueCount,attr" json:"dequeueCount"`
}
