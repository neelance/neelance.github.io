package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"runtime"
	"time"
)

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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	c := make(chan *Repository, 10000)

	go func() {
		from := time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Now()

		for dataTime := to; dataTime.After(from); dataTime = dataTime.Add(-time.Hour) {
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
					panic(err)
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

	db, err := sql.Open("sqlite3", "github.db")
	if err != nil {
		panic(err)
	}

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
