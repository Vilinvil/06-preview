package main

import (
	"context"
	"fmt"
	"log"
	"os"
	pr "thumbnail_utility/api"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = "localhost:50051"

var filePaths = "./test_file_jpg/file%v.jpg"

func isModeAsync(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == "--async"
}

func main() {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("In main did not connect: %v", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("In main can`t close connection")
		}
	}()
	c := pr.NewThumbnailServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var UrlSl []string
	asyncMode := isModeAsync(os.Args[1:])
	if asyncMode {
		UrlSl = os.Args[2:]
	} else {
		UrlSl = os.Args[1:]
	}

	resp, err := c.DownloadThumbnail(ctx, &pr.ThumbnailRequest{Url: UrlSl, Asynchronous: asyncMode})
	if err != nil {
		log.Fatalf("In main could not DownloadThumbnail: %v", err)
	}

	if resp == nil {
		log.Fatalf("In main *preview.ThumbnailResponse == nil")
	}
	for index, img := range resp.Img {
		path := fmt.Sprintf(filePaths, index)
		out, err := os.Create(path)
		if err != nil {
			log.Fatalf("In main could not Create %s: %v", path, err)
		}

		_, err = out.Write(img)
		if err != nil {
			log.Fatalf("In main can`t write in %s: %v", path, err)
		}

		err = out.Close()
		if err != nil {
			log.Printf("In main can`t close %s: %v\n", path, err)
		}
	}

}
