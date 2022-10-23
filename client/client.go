package main

import (
	"context"
	"flag"
	"log"
	"os"
	pr "preview/api"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = "localhost:50051"
)

func isModeAsync(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == "--async"
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("In main can`t close connection")
		}
	}()
	c := pr.NewThumbnailServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.DownloadThumbnail(ctx, &pr.ThumbnailRequest{Url: "https://youtu.be/jfKfPfyJRdk", Asynchronous: isModeAsync(os.Args[1:])})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("DownloadThumbnail len= %v", len(r.Img))

	out, err := os.Create("./test_file_jpg/file.jpg")
	if err != nil {
		log.Println("In main can`t create ./test_file_jpg/file.jpg")
		os.Exit(1)
	}
	defer func() {
		err = out.Close()
		if err != nil {
			log.Println("In main can`t close ./test_file_jpg/file.jpg")
		}
	}()

	_, err = out.Write(r.Img)
	if err != nil {
		log.Println("In main can`t write in file ./test_file_jpg/file.jpg")
		os.Exit(1)
	}

}
