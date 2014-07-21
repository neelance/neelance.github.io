package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
	"time"
)

const interval = 60 * 60 * 24 * 7

func main() {
	db, err := sql.Open("sqlite3", "github.db")
	if err != nil {
		panic(err)
	}

	var languages []string
	languageIndices := make(map[string]int)
	rows, err := db.Query(`select Language from repositories where not Language = "" and Stargazers >= 50 group by Language order by count(*) desc`)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var language string
		if err := rows.Scan(&language); err != nil {
			panic(err)
		}
		languageIndices[language] = len(languages)
		languages = append(languages, language)
		if len(languages) == 20 {
			break
		}
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	start := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Unix() / interval
	end := time.Now().Unix() / interval
	values := make([][]int, end-start)
	for i := range values {
		values[i] = make([]int, len(languages))
	}

	rows, err = db.Query(`select Language, strftime('%s', CreatedAt) / ? as formatted, Count(*) from repositories where Language in ("`+strings.Join(languages, `", "`)+`") and Stargazers >= 50 group by Language, formatted`, interval)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var language string
		var createdAt int64
		var count int
		if err := rows.Scan(&language, &createdAt, &count); err != nil {
			panic(err)
		}
		if createdAt < start {
			createdAt = start
		}
		if createdAt >= end {
			continue
		}
		values[createdAt-start][languageIndices[language]] += count
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	out, err := os.Create("output.tsv")
	if err != nil {
		panic(err)
	}
	out.WriteString("date")
	for _, language := range languages {
		out.WriteString("\t" + language)
	}
	out.WriteString("\n")
	totals := make([]int, len(languages))
	all := 0
	for t, counts := range values {
		// fmt.Println(counts)
		out.WriteString(time.Unix((start+int64(t))*interval, 0).Format("2006-01-02"))
		for i := range languages {
			all += counts[i]
		}
		// fmt.Println(all)
		for i := range languages {
			totals[i] += counts[i]
			fmt.Fprintf(out, "\t%f", float64(totals[i])/float64(all)*100)
		}
		out.WriteString("\n")
	}
	out.Close()

}
