package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

type Repository struct {
	StargazersCount int
	Language        string
	CreatedAt       time.Time
}

func main() {
	repos := make(map[string]*Repository)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	offset := 200000
	for offset >= 100 {
		newOffset := 0
		page := 1
		for page <= 10 {
			time.Sleep(1)
			results, _, err := client.Search.Repositories(context.Background(), fmt.Sprintf("stars:%d..%d", offset/2, offset), &github.SearchOptions{
				Sort:  "stars",
				Order: "desc",
				ListOptions: github.ListOptions{
					PerPage: 100,
					Page:    page,
				},
			})
			if err != nil {
				fmt.Println(err)
				time.Sleep(10 * time.Second)
				continue
			}

			if len(results.Repositories) == 0 {
				fmt.Println("no results")
				break
			}

			for _, repo := range results.Repositories {
				if repo.Language != nil {
					fmt.Println(*repo.StargazersCount, *repo.FullName, *repo.Language)
					repos[*repo.FullName] = &Repository{
						StargazersCount: *repo.StargazersCount,
						Language:        *repo.Language,
						CreatedAt:       repo.CreatedAt.Time,
					}
				}
				newOffset = *repo.StargazersCount
			}

			page++
			fmt.Println("Count:", len(repos))
		}

		offset = newOffset
	}

	file, err := os.Create("../repos.json")
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(file).Encode(repos); err != nil {
		panic(err)
	}
	file.Close()
}
