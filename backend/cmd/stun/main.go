package main

import (
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/stun"
	"github.com/syncloud/redirect/version"
	"go.uber.org/zap"
)

func main() {
	fmt.Printf("redirect stun build=%s sha=%s time=%s\n", version.BuildNumber, version.GitSha, version.BuildTime)
	var address string
	cmd := &cobra.Command{
		Use: "stun",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := log.Default()
			conn, err := net.ListenPacket("udp", address)
			if err != nil {
				return err
			}
			logger.Info("stun listening", zap.String("address", conn.LocalAddr().String()))
			return stun.NewServer(conn, logger).Serve()
		},
	}
	cmd.Flags().StringVar(&address, "address", ":3478", "udp listen address")

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
