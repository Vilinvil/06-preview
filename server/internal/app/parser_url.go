package app

import (
	"fmt"
	"net/url"
)

func GetVideoIdFromUrl(URL string) (string, error) {
	inUrl, err := url.Parse(URL)
	if err != nil {
		return "", fmt.Errorf("in getVideoIdFromUrl can`t parse url: %w", err)
	}

	videoId := inUrl.Query().Get("v")
	if videoId == "" {
		// В старых видео Id лежит не в параметрах, а просто в пути
		if len(inUrl.Path) == 0 {
			return "", fmt.Errorf("in getVideoIdFromUrl len(inUrl.Path) == 0")
		}
		videoId = inUrl.Path[1:]
	}

	return videoId, nil
}
