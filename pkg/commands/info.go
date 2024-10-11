package commands

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ikto/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewInfoCommand() *cobra.Command {
	var socket string
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Print info about the local node",
		RunE: func(cmd *cobra.Command, args []string) error {

			conn, err := grpc.NewClient("0.0.0.0", grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				return net.Dial("unix", socket)
			}), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return err
			}

			client := proto.NewAdminServiceClient(conn)

			infos, err := client.NodeInfo(context.Background(), nil)
			if err != nil {
				return err
			}

			fmt.Println("Local Node:")
			printPeer(infos.Self)
			fmt.Println("Peers:")
			for _, peer := range infos.Peers {
				printPeer(peer)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&socket, "socket", "/tmp/ikto.sock", "Path to the admin socket")

	return cmd
}

func printPeer(peer *proto.Peer) {
	fmt.Printf("Name: %s\n", peer.Name)
	fmt.Printf("Public Key: %s\n", peer.PublicKey)
	fmt.Printf("Advertise Address: %s\n", peer.AdvertiseAddr)
	fmt.Printf("Allowed IP: %s\n", peer.AllowedIp)
	fmt.Printf("WireGuard Port: %d\n", peer.WgPort)
	fmt.Println()
}
