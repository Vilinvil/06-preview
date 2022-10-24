package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	pr "thumbnail_utility/api"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
)

var (
	errPreviewNotFound = errors.New("some preview not found")
	errEmptyUrlSl      = errors.New("UrlSL is empty slice")
	errUnknown         = errors.New("unknown error on server")
)
var maxOldEl = time.Hour

var port = 50051

var dbPath = "../caching_db/preview.db"

var urlYoutubeApi = "https://i1.ytimg.com/vi/%s/maxresdefault.jpg"

const (
	insertSQL = `
INSERT INTO previews (
	url, startTime, file
) VALUES (
	?, ?, ?
)
`

	schemaSQL = `
CREATE TABLE IF NOT EXISTS previews (

url TEXT PRIMARY KEY NOT NULL,

startTime TIMESTAMP,

file BLOB NOT NULL

)`

	selectUrlSQL = `SELECT * FROM previews
    WHERE url == ?`

	deleteElSQL = `
	 DELETE FROM previews
 WHERE startTime <= ?`
)

var db = &DB{}

type ElDB struct {
	URL  string
	Time time.Time
	File []byte
}

type DB struct {
	sql  *sql.DB
	stmt *sql.Stmt
}

func NewDB(dbFile string) (*DB, error) {
	log.Println("in NewDB")
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("in NewDB can`t open file %s: %w", dbFile, err)
	}

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("in NewDB can`t sqlDB.Exec(schemaSQL): %w", err)
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("in NewDB can`t sqlDB.Prepare(insertSQL): %w", err)
	}

	db := DB{
		sql:  sqlDB,
		stmt: stmt,
	}
	return &db, nil
}

func (db *DB) Add(elDB *ElDB) error {
	log.Println("in ADD")
	tx, err := db.sql.Begin()
	if err != nil {
		return fmt.Errorf("in Add can`t begin transaction: %w", err)
	}

	_, err = tx.Stmt(db.stmt).Exec(elDB.URL, elDB.Time, elDB.File)
	if err != nil {
		errIn := tx.Rollback()
		if errIn != nil {
			return fmt.Errorf("in Add can`t tx.Rollback(): %w", err)
		}
		return fmt.Errorf("in Add can`t Exec(): %+v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("in Add can`t tx.Commit(): %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	err := db.stmt.Close()
	if err != nil {
		return fmt.Errorf("in Close() can`t db.stmt.Close(): %w", err)
	}

	err = db.sql.Close()
	if err != nil {
		return fmt.Errorf("in Close() can`t db.sql.Close(): %w", err)
	}

	return nil
}

func (db *DB) Select(url string) (*ElDB, error) {
	log.Println("in select")
	resElDb := &ElDB{}
	row := db.sql.QueryRow(selectUrlSQL, url)

	err := row.Scan(&resElDb.URL, &resElDb.Time, &resElDb.File)
	if err != nil {
		return nil, fmt.Errorf("in Select can`t row.Scan(): %w", err)
	}

	return resElDb, nil
}

func (db *DB) Clearing(period time.Duration) {
	tic := time.NewTicker(period)
	for {
		select {
		case <-tic.C:
			stmt, err := db.sql.Prepare(deleteElSQL)
			if err != nil {
				log.Printf("In Clearing can`t db.sql.Prepare(deleteElSQL): %v", err)
			}
			db.stmt = stmt
			tx, err := db.sql.Begin()
			if err != nil {
				log.Printf("In Clearing can`t db.sql.Begin(): %v", err)
			}

			_, err = tx.Stmt(db.stmt).Exec(time.Now().Add(-1 * maxOldEl))
			if err != nil {
				errIn := tx.Rollback()
				if errIn != nil {
					log.Printf("In Clearing can`t tx.Rollback(): %v", err)
				}
				log.Printf("In Clearing can`t tx.Stmt(db.stmt).Exec(): %v", err)
			}

			err = tx.Commit()
			if err != nil {
				log.Printf("In Clearing can`t tx.Commit(): %v", err)
			}

			stmt, err = db.sql.Prepare(insertSQL)
			if err != nil {
				log.Printf("In Clearing can`t db.sql.Prepare(insertSQL): %v", err)
			}
			db.stmt = stmt

		}
	}
}

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

func (s *server) singleHandler(ctx context.Context, URL string) ([]byte, error) {
	var resImg []byte
	elDb, err := db.Select(URL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("in singleHandler can`t db.Select(): %w", err)
	}
	if elDb != nil {
		if elDb.File == nil {
			return nil, fmt.Errorf("in singleHandler elDb.File == nil")
		}
		if len(elDb.File) == 0 {
			return nil, fmt.Errorf("in singleHandler elDb.File is empty")
		}

		resImg = elDb.File
	} else {
		videoId, err := getVideoIdFromUrl(URL)
		if err != nil {
			return nil, fmt.Errorf("in singleHandler: %w", err)
		}

		resp, err := http.Get(fmt.Sprintf(urlYoutubeApi, videoId))
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t http.Get(): %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusNotFound {
				return nil, errPreviewNotFound
			}
			return nil, fmt.Errorf("in singleHandler http.Get() give unknown error")
		}

		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t Copy resp.Body to buf: %w", err)
		}

		resImg = buf.Bytes()

		err = db.Add(&ElDB{URL: URL, Time: time.Now(), File: resImg})
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t db.Add: %w", err)
		}

		err = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("in singleHandler can`t close resp.body: %w", err)
		}
	}

	return resImg, nil
}

func (s *server) worker(ctx context.Context, in chan string, out chan []byte) {
	for URL := range in {
		resImg, err := s.singleHandler(ctx, URL)
		if err != nil {
			log.Printf("in worker [%d] error: %v", ctx.Value("idWorker"), err)
		}

		out <- resImg
		fmt.Println("worker out")

	}
}

func (s *server) asynchronousHandler(ctx context.Context, UrlSl []string) ([][]byte, error) {
	log.Println("in asynchronousHandler")
	resSl := make([][]byte, 0, len(UrlSl))
	resImg := make([]byte, 0)

	if len(UrlSl) == 0 {
		return nil, errEmptyUrlSl
	}

	wg := &sync.WaitGroup{}
	in := make(chan string)
	out := make(chan []byte)
	for i := 1; i <= 10; i++ {
		ctxWorker := context.WithValue(ctx, "idWorker", i)
		go s.worker(ctxWorker, in, out)
	}
	wg.Add(len(UrlSl))
	go func() {
		for _, URL := range UrlSl {
			in <- URL
			fmt.Println("go in")
		}
	}()

	go func() {
		for {
			select {
			case resImg = <-out:
				fmt.Println("select out")
				resSl = append(resSl, resImg)
				wg.Done()
			}

		}
	}()

	wg.Wait()

	return resSl, nil
}

func (s *server) sequentialHandler(ctx context.Context, UrlSl []string) ([][]byte, error) {
	log.Println("in sequentialHandler")
	resSl := make([][]byte, 0, len(UrlSl))

	if len(UrlSl) == 0 {
		return nil, errEmptyUrlSl
	}

	for _, URL := range UrlSl {
		resImg, err := s.singleHandler(ctx, URL)
		if err != nil {
			return nil, fmt.Errorf("in sequentialHandler can`t s.singleHandler(ctx, URL): %w", err)
		}
		resSl = append(resSl, resImg)
	}

	return resSl, nil
}

func (s *server) DownloadThumbnail(ctx context.Context, in *pr.ThumbnailRequest) (*pr.ThumbnailResponse, error) {
	log.Println("in DownloadThumbnail")
	if in == nil {
		return nil, fmt.Errorf("in DownloadThumbnail *pr.ThumbnailRequest == nil")
	}

	var err error
	resSl := make([][]byte, 0, len(in.Url))
	if in.Asynchronous {
		resSl, err = s.asynchronousHandler(ctx, in.Url)
	} else {
		resSl, err = s.sequentialHandler(ctx, in.Url)
	}

	if err != nil {
		log.Printf("in DownloadThumbnail error is: %w", err)
		switch {
		case errors.Is(err, errPreviewNotFound):
			return nil, errPreviewNotFound
		case errors.Is(err, errEmptyUrlSl):
			return nil, errEmptyUrlSl
		default:
			return nil, errUnknown
		}
	}

	return &pr.ThumbnailResponse{Img: resSl}, nil
}

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("In main failed to listen: %v", err)
	}

	file, err := os.Create(dbPath)
	if err != nil {
		log.Fatalf("In main fail create %s: %v", dbPath, err)
	}
	err = file.Close()
	if err != nil {
		log.Fatalf("In main fail close %s: %v", dbPath, err)
	}

	db, err = NewDB(dbPath)
	if err != nil {
		log.Fatalf("In main fail NewDB() %s: %v", dbPath, err)
	}

	// Каждые 10 мин проверяю, не лежит ли более одного часа ссылка с файлом в базе
	go db.Clearing(10 * time.Minute)

	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("In main fail db.Close(): %v", err)
		}

	}()

	s := grpc.NewServer()
	pr.RegisterThumbnailServiceServer(s, &server{})
	log.Printf("In main server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("In main failed to serve: %v", err)
	}
}
