package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"thumbnail_utility/server/internal/app/server_error"

	"github.com/Vilinvil/preview/pkg/service"
	"google.golang.org/grpc"
	"thumbnail_utility/server/internal/app"
)

func NewListener(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf(":%d", port))
}

type server struct {
	service.UnimplementedThumbnailServiceServer
}

func NewServer() *grpc.Server {
	s := grpc.NewServer()
	service.RegisterThumbnailServiceServer(s, &server{})

	return s
}

func (s *server) DownloadThumbnail(ctx context.Context, in *service.ThumbnailRequest) (*service.ThumbnailResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("in DownloadThumbnail *pr.ThumbnailRequest == nil")
	}

	Url := in.GetUrl()
	if Url == nil {
		return nil, fmt.Errorf("in DownloadThumbnail *pr.ThumbnailRequest.Url == nil")
	}

	var err error
	resSl := make([][]byte, 0, len(in.Url))

	if in.GetAsynchronous() {
		resSl, err = app.AsyncProcess(ctx, Url)
	} else {
		resSl, err = app.SequentialProcess(ctx, Url)
	}

	if err != nil {
		log.Printf("in DownloadThumbnail error is: %v", err)
		switch {
		case errors.Is(err, server_error.ErrPreviewNotFound):
			return nil, server_error.ErrPreviewNotFound
		case errors.Is(err, server_error.ErrEmptyUrlSl):
			return nil, server_error.ErrEmptyUrlSl
		default:
			return nil, server_error.ErrUnknown
		}
	}

	// Добавить пояснения про запрос. Хз какие
	log.Printf("in DownloadThumbnail successfully handle request")
	return &service.ThumbnailResponse{Img: resSl}, nil
}
