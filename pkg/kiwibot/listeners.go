package kiwibot

import (
	"encoding/json"
	"log"
	"net"
)

// StartUDP starts the udp listener
func StartUDP() {
	go listenUDP()
}

func listenUDP() {
	conf := GetConfig()
	addr, _ := net.ResolveUDPAddr("udp", conf.UDPaddr)
	sock, _ := net.ListenUDP("udp", addr)
	sock.SetReadBuffer(20480)

	log.Println("Listening on", addr.String())
	for {
		buf := make([]byte, 1024)
		rlen, _, err := sock.ReadFromUDP(buf)
		if err != nil {
			log.Println("Socket error:", err)
		}
		handlePacket(buf, rlen)
	}
}

func handlePacket(buf []byte, rlen int) {
	if buf[0] != '{' {
		log.Println("Unexpected packet start:", string(buf[0:rlen]))
	}
	var data map[string]interface{}
	err := json.Unmarshal(buf[:rlen], &data)
	if err != nil {
		log.Println("JSON error:", err)
		return
	}
	cmd := data["cmd"].(string)
	switch cmd {
	case "botsend":
		where := GetArray(data["dest"].([]interface{}))
		message := data["msg"].(string)
		msgSend(where, message)
		break
	}
}

func msgSend(where []string, message string) {
	for _, el := range where {
		BotSend(el, message)
	}
}
