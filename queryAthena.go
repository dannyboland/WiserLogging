package main

import (
	"database/sql"
	"fmt"
	_ "github.com/segmentio/go-athena"
	"io/ioutil"
	"log"
	"os"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

func main() {
	query, err := ioutil.ReadFile("athena.sql")
	check(err)
	f, err := os.Create("wiser.tsdb")
	check(err)
	db, _ := sql.Open("athena", "db=your-db&output_location=s3://aws-athena-query-results-youraccountid-eu-west-1&region=eu-west-1")
	_, err = db.Query("MSCK REPAIR TABLE your-db.your-table")
	check(err)
	rows, err := db.Query(fmt.Sprintf("%s", query))
	check(err)

	for rows.Next() {
		var metric string
		var t string
		var value float64
		var tag string
		rows.Scan(&metric, &t, &value, &tag)
		fmt.Fprintln(f, metric, t, value, tag)
	}
}
