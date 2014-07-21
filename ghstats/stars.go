package main

import (
	"compress/gzip"
	// "database/sql"
	"encoding/json"
	"fmt"
	// "github.com/coopernurse/gorp"
	// _ "github.com/mattn/go-sqlite3"
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

// type DbEntry struct {
// 	RepositoryId int
// 	Time         time.Time
// 	Stargazers   int
// }

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// entries := make(chan *DbEntry, 10000)

	// db, err := sql.Open("sqlite3", "./stars.db")
	// if err != nil {
	// 	panic(err)
	// }

	// dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	// dbmap.AddTableWithName(DbEntry{}, "stars")
	// if err := dbmap.CreateTablesIfNotExists(); err != nil {
	// 	panic(err)
	// }

	// go func() {
	// 	for e := range entries {
	// 		if err := dbmap.Insert(e); err != nil {
	// 			panic(err)
	// 		}
	// 	}
	// }()

	dataTime := time.Now()
	for {
		beego := 0
		revel := 0
		for {
			name := fmt.Sprintf("/Volumes/Data/GitHub-Archive/%d-%02d-%02d-%d.json.gz", dataTime.Year(), dataTime.Month(), dataTime.Day(), dataTime.Hour())
			// fmt.Println(name)

			file, err := os.Open(name)
			if err != nil {
				if os.IsNotExist(err) {
					dataTime = dataTime.Add(-time.Hour)
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
				var e Event
				err = dec.Decode(&e)
				if err != nil {
					if err == io.EOF {
						break
					}
					panic(err)
				}
				if e.Repository != nil && e.Repository.Id == 3577919 && beego == 0 {
					beego = e.Repository.Stargazers
				}
				if e.Repository != nil && e.Repository.Id == 2945088 && revel == 0 {
					revel = e.Repository.Stargazers
				}
			}

			gzipReader.Close()
			file.Close()

			if beego != 0 && revel != 0 {
				break
			}
			dataTime = dataTime.Add(-time.Hour)
		}

		fmt.Println(dataTime, beego, revel)
		dataTime = dataTime.Add(-time.Hour * 24 * 7)
	}

}
