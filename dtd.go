// GANGLIA-ACTIVEMQ
// https://github.com/jbuchbinder/ganglia-activemq

package main

type Queues struct {
	Items []Queue `xml:"queue" json:"queue"`
}

type Queue struct {
	Name  string     `xml:"name,attr" json:"name"`
	Stats QueueStats `xml:"stats" json:"stats"`
}

type QueueStats struct {
	Size          int64 `xml:"size,attr" json:"size"`
	ConsumerCount int64 `xml:"consumerCount,attr" json:"consumerCount"`
	EnqueueCount  int64 `xml:"enqueueCount,attr" json:"enqueueCount"`
	DequeueCount  int64 `xml:"dequeueCount,attr" json:"dequeueCount"`
}
