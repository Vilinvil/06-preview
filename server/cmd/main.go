package main

import (
	"log"
	"time"

	"thumbnail_utility/server/internal/app"
	"thumbnail_utility/server/internal/storage/sqlite"
	"thumbnail_utility/server/internal/transport/grpc"
)

const port = 50051

const countOfWorkers = 10

var dbPath = "server/internal/storage/sqlite/cashing_db/thumbnail.db"

func main() {
	chJobs := make(chan app.Job)
	app.InitAsync(chJobs, countOfWorkers)

	db, err := sqlite.NewDB(dbPath)
	if err != nil {
		log.Fatalf("In main can`t create DB: %v", err)
	}
	app.InitDB(db)
	defer func() {
		err = db.Close()
		if err != nil {
			log.Printf("In main fail db.Close(): %v", err)
		}
	}()

	// Каждые 10 мин проверяю, не лежит ли более одного часа ссылка с файлом в базе
	go db.Clearing(10 * time.Minute)

	lis, err := grpc.NewListener(port)
	if err != nil {
		log.Fatalf("In main failed to listen: %v", err)
	}

	server := grpc.NewServer()

	log.Printf("In main server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("In main failed to serve: %v", err)
	}
}
