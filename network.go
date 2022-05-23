package cloudy

import (
	"net"
	"net/http"
	"time"
)

func CheckAddress(address string, timeout time.Duration) bool {
	_, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	} else {
		return true
	}
}

func WaitForAddress(address string, timeout time.Duration) bool {
	end := time.Now().Add(timeout)
	for {
		if time.Now().After(end) {
			return false
		}
		_, err := http.Get(address)
		if err == nil {
			return true
		}
	}
}
