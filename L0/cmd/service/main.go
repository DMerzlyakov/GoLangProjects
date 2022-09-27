package main

import (
	"L0/internal/server"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go server.StartAPI()
	server.ListenToNutsStreaming()
	wg.Wait()
}
