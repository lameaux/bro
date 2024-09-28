package grpcserver

import (
	"fmt"
	"net"

	pb "github.com/lameaux/bro/protos/metrics"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMetricsServer
}

func StartGrpcServer(port int) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	newServer := grpc.NewServer()
	pb.RegisterMetricsServer(newServer, &server{})

	go func() {
		if err = newServer.Serve(lis); err != nil {
			log.Fatal().Err(err).
				Int("port", port).
				Msg("failed to start grpc server")
		}
	}()

	log.Debug().Int("port", port).Msg("grpc server started")

	return newServer, nil
}

func StopGrpcServer(s *grpc.Server) {
	s.GracefulStop()

	log.Debug().Msg("grpc server stopped")
}