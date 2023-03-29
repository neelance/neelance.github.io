package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

func main() {
	f, err := os.Create("../repos.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	n := 0
	defer w.Flush()
	w.Write([]string{"name", "createdAt", "language", "stargazers"})

	offset := 500000
	for offset >= 1000 {
		var after *string
		newOffset := 0
		for {
			var data struct {
				Search struct {
					PageInfo struct {
						HasNextPage bool
						EndCursor   string
					}
					Nodes []struct {
						NameWithOwner string
						CreatedAt     time.Time
						Stargazers    struct {
							TotalCount int
						}
						PrimaryLanguage struct {
							Name string
						}
					}
				}
			}

			graphqlRequest(`
				query ($query: String!, $after: String) {
					search(type: REPOSITORY, query: $query, first: 100, after: $after) {
						pageInfo {
							hasNextPage
							endCursor
						}
						repositoryCount
						nodes {
							... on Repository {
								nameWithOwner
								createdAt
								stargazers {
									totalCount
								}
								primaryLanguage {
									name
								}
							}
						}
					}
				}
			`, map[string]interface{}{
				"query": fmt.Sprintf("stars:%d..%d", offset/2, offset),
				"after": after,
			}, &data)

			for _, repo := range data.Search.Nodes {
				fmt.Println(repo.Stargazers.TotalCount, repo.NameWithOwner, repo.PrimaryLanguage.Name)
				w.Write([]string{
					repo.NameWithOwner,
					repo.CreatedAt.Format("2006-01-02"),
					repo.PrimaryLanguage.Name,
					strconv.Itoa(repo.Stargazers.TotalCount),
				})
				n++
				newOffset = repo.Stargazers.TotalCount
			}
			fmt.Println("Count:", n)

			if !data.Search.PageInfo.HasNextPage {
				break
			}
			after = &data.Search.PageInfo.EndCursor
		}

		offset = newOffset
	}
}

var client = oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(
	&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
))

func graphqlRequest(query string, variables map[string]interface{}, data interface{}) {
	body, err := json.Marshal(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		panic(err)
	}

	for {
		resp, err := client.Post("https://api.github.com/graphql", "application/json", bytes.NewReader(body))
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()

		respBody := struct {
			Data   interface{} `json:"data"`
			Errors []struct {
				Message string
			}
		}{
			Data: data,
		}

		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			panic(err)
		}

		if len(respBody.Errors) != 0 {
			for _, err := range respBody.Errors {
				log.Println(err.Message)
			}
			continue
		}

		break
	}
}
