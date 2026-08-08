package main

import (
	"flag"
	"log"
	"net"
	"strconv"
)

func main() {
	smtpAddress := flag.String("smtp", "127.0.0.1:2525", "smtp address the tunnel forwards to")
	apiAddress := flag.String("api", ":4580", "control api address")
	frpc := flag.String("frpc", "/usr/local/bin/frpc", "frpc binary")
	serverAddr := flag.String("server-addr", "", "frps address")
	serverName := flag.String("server-name", "", "frps tls server name")
	workDir := flag.String("work-dir", "/tmp/device-faker", "where tunnel configs and logs go")
	flag.Parse()

	_, port, err := net.SplitHostPort(*smtpAddress)
	if err != nil {
		log.Fatalf("smtp address is not host:port: %v", err)
	}
	localPort, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("smtp port is not a number: %v", err)
	}

	mailbox := NewMailbox()
	tunnels := NewTunnels(*frpc, *serverAddr, *serverName, localPort, *workDir)

	if err := NewSmtpServer(*smtpAddress, mailbox).Start(); err != nil {
		log.Fatalf("device smtp failed to start: %v", err)
	}
	if err := NewRest(*apiAddress, mailbox, tunnels).Start(); err != nil {
		log.Fatalf("device api failed to start: %v", err)
	}
	log.Printf("device faker smtp %s api %s", *smtpAddress, *apiAddress)
	select {}
}
