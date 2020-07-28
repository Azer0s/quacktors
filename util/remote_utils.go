package util

import (
	"encoding/json"
	"github.com/Azer0s/quacktors/messages"
	"net"
	"regexp"
	"strconv"
)

var addrRegex *regexp.Regexp

//noinspection RegExpRedundantEscape
func init() {
	r, err := regexp.Compile("^(\\w+)@((?:\\w+)|(?:\\w+\\.\\w+)|(?:\\d+\\.\\d+\\.\\d+\\.\\d+)|(?:\\[(?:[\\dA-Fa-f]*:?)+\\])|):(\\d+)$")

	if err != nil {
		panic(err)
	}

	addrRegex = r
}

func ParseAddress(addr string) (system, address string, port int, err error) {
	if !addrRegex.MatchString(addr) {
		return "", "", 0, InvalidAddressError()
	}

	matches := addrRegex.FindStringSubmatch(addr)
	system = matches[1]
	address = matches[2]
	p, err := strconv.Atoi(matches[3])

	if err != nil {
		return "", "", 0, err
	}

	port = p
	return
}

func SendErr(connection *net.UDPConn, addr *net.UDPAddr) {
	b, _ := json.Marshal(&messages.GatewayResponse{Err: true})
	_, _ = connection.WriteToUDP(b, addr)
}
