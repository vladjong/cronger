package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/vladjong/cronger"
)

const (
	dns = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
)

func main() {
	// Connect sqlx
	db, err := sqlx.Connect("postgres", dns)
	if err != nil {
		log.Fatalln(err)
	}

	// Init cronger
	cr, err := cronger.New(&cronger.Config{
		Loc:        time.UTC,
		TypeClient: cronger.Sqlx,
		Client:     db,
		IsMigrate:  true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Start async cronger
	cr.StartAsync()

	// View backup jobs
	backJob := cr.BackupJobs()
	fmt.Println(backJob)

	// Add jobs
	if err := cr.AddJob("title1", "tag1", "* * * * *", func() {
		fmt.Println("done task 1")
	}); err != nil {
		log.Fatalln(err)
	}

	if err := cr.AddJob("title2", "tag2", "* * * * *", func() {
		fmt.Println("done task 2")
	}); err != nil {
		log.Fatalln(err)
	}

	if err := cr.AddJob("title3", "tag3", "* * * * *", func() {
		fmt.Println("done task 3")
	}); err != nil {
		log.Fatalln(err)
	}

	// View jobs
	job, err := cr.Jobs()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(job)

	time.Sleep(time.Minute)

	backJob = cr.BackupJobs()
	fmt.Println(backJob)

	// Delete jobs
	cr.RemoveJob("tag2")

	job, err = cr.Jobs()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(job)

	time.Sleep(time.Minute * 5)
}
