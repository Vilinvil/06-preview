package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	pr "preview/api"
)

var errPreviewNotFound = errors.New("preview not found")

var port = flag.Int("port", 50051, "The server port")

var urlYoutubeApi = "https://i1.ytimg.com/vi/%s/maxresdefault.jpg"

func getVideoIdFromUrl(URL string) (string, error) {
	inUrl, err := url.Parse(URL)
	if err != nil {
		return "", fmt.Errorf("in getVideoIdFromUrl can`t parse url: %w", err)
	}

	videoId := inUrl.Query().Get("v")
	if videoId == "" {
		// В старых видео Id лежит не в параметрах, а просто в пути
		videoId = inUrl.Path[1:]
	}

	return videoId, nil
}

type server struct {
	pr.UnimplementedThumbnailServiceServer
}

func (c *server) DownloadThumbnail(ctx context.Context, in *pr.ThumbnailRequest) (*pr.ThumbnailResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("in DownloadThumbnail *pr.ThumbnailRequest == nil")
	}

	videoId, err := getVideoIdFromUrl(in.Url)
	if err != nil {
		return nil, fmt.Errorf("in DownloadThumbnail: %w", err)
	}

	resp, err := http.Get(fmt.Sprintf(urlYoutubeApi, videoId))
	if err != nil {
		return nil, fmt.Errorf("in DownloadThumbnail can`t Get: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Println("In DownloadThumbnail can`t close resp.body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errPreviewNotFound
		}
		return nil, fmt.Errorf("unknown error")
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("in DownloadThumbnail can`t Copy: %w", err)
	}

	return &pr.ThumbnailResponse{Img: buf.Bytes()}, nil
}
func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pr.RegisterThumbnailServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
