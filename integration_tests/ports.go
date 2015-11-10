package integration_tests

import (
	"sync"
)

var port int = 10000
var l *sync.Mutex = &sync.Mutex{}

func GetUnusedPort() int {
	l.Lock()
	defer l.Unlock()
	port = port + 1
	return port - 1
}
