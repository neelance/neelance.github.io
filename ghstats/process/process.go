package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

const interval = 60 * 60 * 24 * 7

type Repository struct {
	Language  string
	CreatedAt time.Time
}

func main() {
	file, err := os.Open("../repos.csv")
	if err != nil {
		panic(err)
	}
	var repos []*Repository
	r := csv.NewReader(file)
	r.Read() // skip headers
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		t, err := time.Parse("2006-01-02", record[1])
		if err != nil {
			panic(err)
		}
		repos = append(repos, &Repository{
			Language:  record[2],
			CreatedAt: t,
		})
	}
	file.Close()

	languageCounts := make(map[string]int)
	for _, repo := range repos {
		languageCounts[repo.Language]++
	}

	fmt.Println(languageCounts)

	var languages []string
	for lang, count := range languageCounts {
		if count > 100 && lang != "" {
			languages = append(languages, lang)
		}
	}
	sort.Strings(languages)
	languageIndices := make(map[string]int)
	for i, lang := range languages {
		languageIndices[lang] = i
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
	totals := make([]float64, len(languages))
	for t, counts := range values {
		out.WriteString(time.Unix((start+int64(t))*interval, 0).Format("2006-01-02"))
		for i := range languages {
			totals[i] += float64(counts[i])
			// totals[i] = totals[i] * 0.99
		}
		all := float64(0)
		for i := range languages {
			all += totals[i]
		}
		for i := range languages {
			v := totals[i] / all * 100
			if v < 0.1 {
				v = 0.1
			}
			fmt.Fprintf(out, "\t%f", v)
		}
		out.WriteString("\n")
	}
	out.Close()
}
