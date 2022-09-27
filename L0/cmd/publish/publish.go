package main

import (
	"github.com/nats-io/stan.go"
	"io/ioutil"
	"os"
)

func main() {
	publish()
}
func publish() {
	conn, _ := stan.Connect("test-cluster", "pub", stan.NatsURL("0.0.0.0:4223"))
	defer conn.Close()
	file, _ := os.Open("cmd/publish/model.json")

	data, _ := ioutil.ReadAll(file)
	_ = conn.Publish("test", data)
	println("successfully publish")
}

//docker run --rm -ti -p4223:4222 nats-streaming
