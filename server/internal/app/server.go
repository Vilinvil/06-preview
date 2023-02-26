package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"thumbnail_utility/server/internal/app/server_error"
	"thumbnail_utility/server/internal/storage/sqlite"
	"thumbnail_utility/server/internal/transport/http"
)

var globChJobs = make(chan Job)

var db = &sqlite.DB{}

type Job struct {
	url string
	out chan jobResult
}

type jobResult struct {
	file []byte
	err  error
}

func worker(ctx context.Context) {
	for curJob := range globChJobs {
		resImg, err := singleHandler(ctx, curJob.url)
		if err != nil {
			log.Printf("in worker [%d] error: %v", ctx.Value("idWorker"), err)
		}

		curJob.out <- jobResult{file: resImg, err: err}
	}
}

func InitDB(DB *sqlite.DB) {
	db = DB
}

func InitAsync(chJobs chan Job, countGoroutine int) {
	globChJobs = chJobs
	for i := 0; i < countGoroutine; i++ {
		ctxWorker := context.WithValue(context.Background(), "idWorker", i)
		go worker(ctxWorker)
	}
}

func singleHandler(ctx context.Context, URL string) ([]byte, error) {
	var resImg []byte
	thumbnailModel, err := db.Select(URL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("in singleHandler can`t db.Select(): %w", err)
	}
	if thumbnailModel != nil {
		if thumbnailModel.File == nil {
			return nil, fmt.Errorf("in singleHandler thumbnailModel.File == nil")
		}
		if len(thumbnailModel.File) == 0 {
			return nil, fmt.Errorf("in singleHandler thumbnailModel.File is empty")
		}

		resImg = thumbnailModel.File
	} else {
		videoId, err := GetVideoIdFromUrl(URL)
		if err != nil {
			return nil, fmt.Errorf("in singleHandler: %w", err)
		}

		resImg, err = http.GetImg(videoId)
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t http.Get(): %w", err)
		}

		err = db.Add(&sqlite.PreviewTableModel{URL: URL, Time: time.Now(), File: resImg})
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t db.Add: %w", err)
		}

	}

	return resImg, nil
}

func AsyncProcess(ctx context.Context, UrlSl []string) ([][]byte, error) {
	resSl := make([][]byte, 0, len(UrlSl))
	jobRes := jobResult{file: make([]byte, 0)}

	if len(UrlSl) == 0 {
		return nil, fmt.Errorf("in asyncProcess: %w", server_error.ErrEmptyUrlSl)

	}

	wg := &sync.WaitGroup{}
	wg.Add(len(UrlSl))
	out := make(chan jobResult)
	go func() {
		for _, URL := range UrlSl {
			curJob := Job{url: URL, out: out}
			globChJobs <- curJob
		}
	}()

	go func() {
		for {
			select {
			case jobRes = <-out:
				if jobRes.err != nil {
					log.Printf("in asyncProcess in gorutine: %v", jobRes.err)
					break
				}
				resSl = append(resSl, jobRes.file)

			}
			wg.Done()
		}
	}()

	wg.Wait()
	if len(UrlSl) > len(resSl) {
		return nil, fmt.Errorf("len(UrlSl) > len(resSl)")
	}

	return resSl, nil
}

func SequentialProcess(ctx context.Context, UrlSl []string) ([][]byte, error) {
	resSl := make([][]byte, 0, len(UrlSl))

	if len(UrlSl) == 0 {
		return nil, fmt.Errorf("in sequentialProcess: %w", server_error.ErrEmptyUrlSl)
	}

	for _, URL := range UrlSl {
		resImg, err := singleHandler(ctx, URL)
		if err != nil {
			return nil, fmt.Errorf("in sequentialProcess can`t s.singleHandler(ctx, URL): %w", err)
		}
		resSl = append(resSl, resImg)
	}

	return resSl, nil
}
