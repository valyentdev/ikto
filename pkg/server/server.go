package server

import (
	"context"
	"net"

	"github.com/valyentdev/ikto/pkg/ikto"
	"github.com/valyentdev/ikto/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func StartAdminServer(ctx context.Context, ikto ikto.Ikto, socket string) error {
	s := &server{
		ikto: ikto,
	}

	ln, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	proto.RegisterAdminServiceServer(grpcServer, s)

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(ln); err != nil {
		return err
	}

	return nil
}

type server struct {
	ikto ikto.Ikto
}

// NodeInfo implements proto.AdminServiceServer.
func (s *server) NodeInfo(context.Context, *emptypb.Empty) (*proto.NodeInfoResponse, error) {
	self := s.ikto.Self()
	peers := s.ikto.Peers()

	peersProto := make([]*proto.Peer, 0, len(peers))

	for _, peer := range peers {
		peersProto = append(peersProto, &proto.Peer{
			Name:          peer.Name,
			PublicKey:     peer.PublicKey.String(),
			AdvertiseAddr: peer.AdvertiseAddress,
			AllowedIp:     peer.AllowedIP,
			WgPort:        int32(peer.WGPort),
		})
	}

	return &proto.NodeInfoResponse{
		Self: &proto.Peer{
			Name:          self.Name,
			PublicKey:     self.PublicKey.String(),
			AdvertiseAddr: self.AdvertiseAddress,
			AllowedIp:     self.AllowedIP,
			WgPort:        int32(self.WGPort),
		},
		Peers: peersProto,
	}, nil

}

var _ proto.AdminServiceServer = (*server)(nil)
