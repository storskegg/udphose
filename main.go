package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	addr string
	port string
)

var (
	cmdRoot = cobra.Command{
		Use:   "udphose",
		Short: "firehose UDP packets to designated host",
		RunE:  execRoot,
	}
)

func execRoot(cmd *cobra.Command, args []string) error {
	ap := addr + ":" + port
	if ap == ":" {
		cmd.Usage()
		os.Exit(0)
	}

	udpServer, err := net.ResolveUDPAddr("udp", ap)
	if err != nil {
		println("ResolveUDPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, udpServer)
	if err != nil {
		println("Listen failed:", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sendData := ""
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				i++
				sendData = "Packet count " + strconv.Itoa(i)
				_, err = conn.Write([]byte(sendData))
				if err != nil {
					fmt.Println("Write data failed:", err.Error())
					os.Exit(1)
				}
				if i%100000 == 0 {
					fmt.Println("Sent :", sendData)
				}
				time.Sleep(time.Microsecond)
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigChan:
				cancel()
				return
			}
		}
	}()

	return nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(100)
		}
	}()

	cmdRoot.Flags().StringVarP(&addr, "addr", "a", "", "UDP address to target")
	cmdRoot.Flags().StringVarP(&port, "port", "p", "", "UDP port to target")

	if err := cmdRoot.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
