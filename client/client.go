package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	pr "thumbnail_utility/api"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var addr = "localhost:50051"

var (
	dirPath   = "./test_file_jpg"
	filePaths = "/file%v.jpg"
)

func isModeAsync(args []string) bool {
	if len(args) == 0 {
		return false
	}
	return args[0] == "--async"
}

func splitStr(sl []string) []string {
	var res []string
	for _, val := range sl {
		res = append(res, strings.Fields(val)...)
	}

	return res
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var UrlSl []string
	asyncMode := isModeAsync(os.Args[1:])
	if asyncMode {
		UrlSl = os.Args[2:]
	} else {
		UrlSl = os.Args[1:]
	}
	UrlSl = splitStr(UrlSl)

	resp, err := c.DownloadThumbnail(ctx, &pr.ThumbnailRequest{Url: UrlSl, Asynchronous: asyncMode})
	if err != nil {
		log.Fatalf("In main could not DownloadThumbnail: %v", err)
	}

	if resp == nil {
		log.Fatalf("In main *preview.ThumbnailResponse == nil")
	}
	for index, img := range resp.Img {
		err = os.Mkdir(dirPath, 0777)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("In main could not Mkdir %s: %v", dirPath, err)
		}
		path := fmt.Sprintf(dirPath+filePaths, index)
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
