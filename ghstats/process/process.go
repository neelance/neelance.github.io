package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const interval = 60 * 60 * 24 * 7

type Repository struct {
	StargazersCount int
	Language        string
	CreatedAt       time.Time
}

func main() {
	file, err := os.Open("../repos.json")
	if err != nil {
		panic(err)
	}

	var repos map[string]*Repository
	if err := json.NewDecoder(file).Decode(&repos); err != nil {
		panic(err)
	}
	file.Close()

	languageCounts := make(map[string]int)
	for _, repo := range repos {
		languageCounts[repo.Language]++
	}

	fmt.Println(languageCounts)

	var languages []string
	languageIndices := make(map[string]int)
	for lang, count := range languageCounts {
		if count > 200 {
			languageIndices[lang] = len(languages)
			languages = append(languages, lang)
		}
	}

	fmt.Println(languages, len(languages))

	start := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC).Unix() / interval
	end := time.Now().Unix() / interval
	values := make([][]int, end-start)
	for i := range values {
		values[i] = make([]int, len(languages))
	}

	for _, repo := range repos {
		createdAt := repo.CreatedAt.Unix() / interval
		if createdAt < start {
			createdAt = start
		}
		if createdAt >= end {
			continue
		}
		if langIndex, ok := languageIndices[repo.Language]; ok {
			values[createdAt-start][langIndex]++
		}
	}

	out, err := os.Create("../output.tsv")
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
		out.WriteString(time.Unix((start+int64(t))*interval, 0).Format("2006-01-02"))
		for i := range languages {
			all += counts[i]
		}
		for i := range languages {
			totals[i] += counts[i]
			fmt.Fprintf(out, "\t%f", float64(totals[i])/float64(all)*100)
		}
		out.WriteString("\n")
	}
	out.Close()
}
