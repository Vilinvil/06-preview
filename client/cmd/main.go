package main

import (
	"context"
	"log"
	"time"

	"thumbnail_utility/client/internal/storage/local"
	"thumbnail_utility/client/internal/transport/grpc"
	"thumbnail_utility/client/internal/utility"
)

var addr = "localhost:50051"

func main() {
	UrlSl, asyncMode := utility.ParseArguments()
	if UrlSl == nil || len(UrlSl) == 0 {
		log.Fatalf("In main UrlSl is uncorrect")
	}

	chCloseConn := make(chan struct{})
	client, err := grpc.NewClient(addr, chCloseConn)
	if err != nil {
		log.Fatalf("In main can`t create grpc client: %v", err)
	}
	defer func() {
		chCloseConn <- struct{}{}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := grpc.DownloadThumbnail(ctx, grpc.RequestThumbnail{Client: client, UrlSl: UrlSl, AsyncMode: asyncMode})
	if err != nil {
		log.Fatalf("In main could not DownloadThumbnail: %v", err)
	}
	if resp == nil {
		log.Fatalf("In main resp == nil")
	}

	err = local.SaveImg(resp.Img)
	if err != nil {
		log.Fatalf("In main: %v", err)
	}

}
