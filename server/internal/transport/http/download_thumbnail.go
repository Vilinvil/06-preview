package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"thumbnail_utility/server/internal/app/server_error"
)

var urlYoutubeApi = "https://i1.ytimg.com/vi/%s/maxresdefault.jpg"

func GetImg(videoId string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf(urlYoutubeApi, videoId))
	if err != nil {
		return nil, fmt.Errorf("in singleHandler can`t http.Get(): %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("in singleHandler url = %v: %w", fmt.Sprintf(urlYoutubeApi, videoId), server_error.ErrPreviewNotFound)
		}
		return nil, fmt.Errorf("in singleHandler http.Get() give unknown error")
	}

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("in singleHandler can`t Copy resp.Body to buf: %w", err)
	}

	resImg := buf.Bytes()

	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("in singleHandler can`t close resp.body: %w", err)
	}

	return resImg, nil
}
