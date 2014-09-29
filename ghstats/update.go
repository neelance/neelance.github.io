package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

var start = time.Date(2014, 9, 27, 0, 0, 0, 0, time.UTC)
var end = time.Now()
var db *sql.DB

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var err error
	db, err = sql.Open("sqlite3", os.Getenv("GOPATH")+"/src/github.com/neelance/neelance.github.io/ghstats/github.db")
	if err != nil {
		panic(err)
	}

	download()
	updateDB()
	collectStats()
}

func download() {
	for dataTime := end; dataTime.After(start); dataTime = dataTime.Add(-time.Hour) {
		name := fmt.Sprintf("%d-%02d-%02d-%d.json.gz", dataTime.Year(), dataTime.Month(), dataTime.Day(), dataTime.Hour())
		dir := fmt.Sprintf("/Volumes/Data/GitHub-Archive/%d-%02d", dataTime.Year(), dataTime.Month())
		fmt.Print(name + " ")

		_, err := os.Stat(dir + "/" + name)
		if err == nil {
			fmt.Println("exists")
			continue
		}
		resp, _ := http.Get("http://data.githubarchive.org/" + name)
		fmt.Println(resp.Status)
		if resp.StatusCode == http.StatusOK {
			os.Mkdir(dir, 0777)
			file, err := os.Create(dir + "/" + name)
			if err != nil {
				panic(err)
			}
			io.Copy(file, resp.Body)
			file.Close()
			resp.Body.Close()
		}
	}
}

type Event struct {
	CreatedAt  time.Time   `json:"created_at"`
	Repository *Repository `json:"repository"`
}

type Repository struct {
	Id           int       `json:"id"`
	Organization string    `json:"organization"`
	Owner        string    `json:"owner"`
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Fork         bool      `json:"fork"`
	Forks        int       `json:"forks"`
	CreatedAt    time.Time `json:"created_at"`
	PushedAt     time.Time `json:"pushed_at"`
	LastEventAt  time.Time `json:"-"`
	Description  string    `json:"description"`
	Homepage     string    `json:"homepage"`
	Language     string    `json:"language"`
	MasterBranch string    `json:"master_branch"`
	Watchers     int       `json:"watchers"`
	Stargazers   int       `json:"stargazers"`
	OpenIssues   int       `json:"open_issues"`
	Size         int       `json:"size"`
	HasIssues    bool      `json:"has_issues"`
	HasWiki      bool      `json:"has_wiki"`
	HasDownloads bool      `json:"has_downloads"`
}

func updateDB() {
	c := make(chan *Repository, 10000)

	go func() {
		for dataTime := end; dataTime.After(start); dataTime = dataTime.Add(-time.Hour) {
			name := fmt.Sprintf("/Volumes/Data/GitHub-Archive/%d-%02d/%d-%02d-%02d-%d.json.gz", dataTime.Year(), dataTime.Month(), dataTime.Year(), dataTime.Month(), dataTime.Day(), dataTime.Hour())
			fmt.Println(name)

			file, err := os.Open(name)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("not found")
					continue
				}
				panic(err)
			}
			gzipReader, err := gzip.NewReader(file)
			if err != nil {
				panic(err)
			}
			dec := json.NewDecoder(gzipReader)

			for {
				// var i interface{}
				// err = dec.Decode(&i)
				// fmt.Println(i)
				// return

				var e Event
				err = dec.Decode(&e)
				if err != nil {
					if err == io.EOF {
						break
					}
					fmt.Println("could not decode entry")
					continue
				}
				if e.Repository != nil {
					e.Repository.LastEventAt = e.CreatedAt
					c <- e.Repository
				}
			}

			gzipReader.Close()
			file.Close()
		}
		close(c)
	}()

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	dbmap.AddTableWithName(Repository{}, "repositories").SetKeys(false, "Id")
	if err := dbmap.CreateTablesIfNotExists(); err != nil {
		panic(err)
	}

	lastEvent := make([]time.Time, 30000000)
	newCount := 0
	updatedCount := 0

	go func() {
		for {
			fmt.Printf("%d new, %d updated\n", newCount, updatedCount)
			newCount = 0
			updatedCount = 0
			time.Sleep(time.Second)
		}
	}()

	for r := range c {
		if lastEvent[r.Id].After(r.LastEventAt) {
			continue
		}
		if e, _ := dbmap.Get(Repository{}, r.Id); e != nil {
			existing := e.(*Repository)
			if r.LastEventAt.After(existing.LastEventAt) {
				lastEvent[r.Id] = r.LastEventAt
				if _, err := dbmap.Update(r); err != nil {
					panic(err)
				}
				updatedCount++
				continue
			}
			lastEvent[r.Id] = existing.LastEventAt
			continue
		}
		lastEvent[r.Id] = r.LastEventAt
		if err := dbmap.Insert(r); err != nil {
			panic(err)
		}
		newCount++
	}
}

const interval = 60 * 60 * 24 * 7

func collectStats() {
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
