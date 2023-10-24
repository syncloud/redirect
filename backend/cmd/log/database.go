package main

import (
	"database/sql"
	"fmt"
	"github.com/klauspost/compress/zip"
	_ "github.com/lib/pq"
	"os"
	"time"

	"log"
)

type Database struct {
	db *sql.DB
}

func (d *Database) Connect() {
	connStr := "user=route53 dbname=route53 sslmode=disable password=pass"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	d.db = db
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) Insert(timestamp, domain, resolver string) {
	insertStatement := `
		INSERT INTO request (timestamp, domain, resolver)
		VALUES ($1, $2, $3)
	`

	t, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.Fatal(err)
	}
	_, err = d.db.Exec(insertStatement, t, domain, resolver)
	if err != nil {
		fmt.Println("Error inserting record:", err)
		return
	}

}

func (d *Database) ProcessFile(fp string, fi os.DirEntry, err error) error {
	if err != nil {
		fmt.Println(err) // can't walk here,
		return nil       // but continue walking elsewhere
	}
	if fi.IsDir() {
		return nil // not a file. ignore.
	}
	fmt.Println(fp)

	zipFile, err := zip.OpenReader(zipFilePath)
	if err != nil {
		fmt.Println("Error opening ZIP file:", err)
		return
	}
	defer zipFile.Close()

	return nil
}

/*	rows, err := db.Query("SELECT timestamp, domain, resolver FROM request")
	if err != nil {
		fmt.Println("Error executing query:", err)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		fmt.Println("No rows returned")
	} else {
		var timestamp time.Time
		var domain, resolver string
		if err := rows.Scan(&timestamp, &domain, &resolver); err != nil {
			fmt.Println("Error scanning rows:", err)
			return
		}
		fmt.Printf("%s: %s %s", timestamp, domain, resolver)
	}*/
