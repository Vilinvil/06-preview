package grpc

import (
	"context"
	"fmt"
	"log"

	"github.com/Vilinvil/preview/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RequestThumbnail struct {
	Client    service.ThumbnailServiceClient
	UrlSl     []string
	AsyncMode bool
}

func NewClient(addr string, chClose chan struct{}) (service.ThumbnailServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		err = fmt.Errorf("in NewClient did not connect %w", err)
		return nil, err
	}
	if chClose == nil {
		return nil, fmt.Errorf("in NewClient chClose == nil")
	}

	go func() {
		<-chClose
		err = conn.Close()
		if err != nil {
			log.Println("In main can`t close connection")
		}
	}()

	return service.NewThumbnailServiceClient(conn), nil
}

func DownloadThumbnail(ctx context.Context, req RequestThumbnail) (*service.ThumbnailResponse, error) {
	if req.Client == nil {
		return nil, fmt.Errorf("in DownloadThumbnail req.Client == nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("in DownloadThumbnail ctx == nil")
	}

	return req.Client.DownloadThumbnail(ctx, &service.ThumbnailRequest{Url: req.UrlSl, Asynchronous: req.AsyncMode})
}
