package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	for dataTime := time.Now(); dataTime.After(time.Date(2012, 2, 1, 0, 0, 0, 0, time.UTC)); dataTime = dataTime.Add(-time.Hour) {
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
