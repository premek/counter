package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

var svgTemplate = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="80" height="20" viewBox="0 0 20 5" version="1.1" xmlns="http://www.w3.org/2000/svg" >
<rect style="fill:#000000;" id="rect111" width="20" height="5" x="0" y="0" />
<text style="font-weight:bold;font-size:4px;line-height:1.25;font-family:monospace;letter-spacing:0px;fill:#00ff00;text-anchor:end;text-align:end" x="18" y="4">
%07d
</text>
</svg>
`

var db *sql.DB
var insertStmt *sql.Stmt
var query *sql.Stmt

func initDb() {
	db, err := sql.Open("sqlite3", "./counter.sqlite3")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	 CREATE TABLE IF NOT EXISTS counter (
         timestamp INTEGER NOT NULL,
         site TEXT);
     CREATE INDEX IF NOT EXISTS i_counter_site ON counter(site);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}

	insertStmt, err = db.Prepare("insert into counter(timestamp, site) values(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	query, err = db.Prepare("select count(*) from counter where site=?")

	if err != nil {
		fmt.Printf("%s", err)
	}

	// TODO where to defer-close db and prepared statements?
}

func increase(site string) {
	_, err := insertStmt.Exec(time.Now().Unix(), site)
	if err != nil {
		log.Fatal(err)
	}
}

func getCount(site string) int {
	var count int
	err := query.QueryRow(site).Scan(&count)

	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		log.Fatal(err)
	}
	return count
}

func handle(w http.ResponseWriter, r *http.Request) {
	site := chi.URLParam(r, "site")
	increase(site) // TODO add visitor ID - hash of IP?
	count := getCount(site)
	svg := fmt.Sprintf(svgTemplate, count)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svg))
}

func main() {
	initDb()
	r := chi.NewRouter()
	r.Get("/{site}.svg", handle)
	http.ListenAndServe(":3000", r)
}
