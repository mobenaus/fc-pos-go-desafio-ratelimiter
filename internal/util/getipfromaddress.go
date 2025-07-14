package util

import (
	"strings"
)

func GetIpFromAddress(remoteAddr string) string {
	sa := strings.Split(remoteAddr, ":")
	ip := sa[:len(sa)-1]
	return strings.Join(ip, ":")
}
