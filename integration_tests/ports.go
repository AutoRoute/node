package integration_tests

import (
	"sync"
)

var port int = 10000
var l *sync.Mutex = &sync.Mutex{}

func GetUnusedPort() int {
	l.Lock()
	defer l.Unlock()
	ret := port
	port = port + 1
	return ret
}
