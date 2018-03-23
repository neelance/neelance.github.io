package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
)

type Repository struct {
	StargazersCount int
	Language        string
	CreatedAt       time.Time
}

func main() {
	repos := make(map[string]*Repository)

	offset := 200000
	for offset >= 200 {
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
				repos[repo.NameWithOwner] = &Repository{
					StargazersCount: repo.Stargazers.TotalCount,
					Language:        repo.PrimaryLanguage.Name,
					CreatedAt:       repo.CreatedAt,
				}
				newOffset = repo.Stargazers.TotalCount
			}
			fmt.Println("Count:", len(repos))

			if !data.Search.PageInfo.HasNextPage {
				break
			}
			after = &data.Search.PageInfo.EndCursor
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

	resp, err := client.Post("https://api.github.com/graphql", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
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
			fmt.Println(err.Message)
		}
		panic("graphql errors")
	}
}
