package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	title    string
	company  string
	location string
	id       string
}

//Scrapper scrapes jobs from kr.indeed.com
func Scrapper(keyword string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + keyword + "&limit=50"
	var jobs []extractedJob

	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(baseURL, i, c)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)
	fmt.Println("Found " + strconv.Itoa(len(jobs)) + " jobs")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"title", "company", "location", "link"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.title, job.company, job.location, "https://kr.indeed.com/viewjob?jk=" + job.id}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}
}

func getPage(url string, page int, mainC chan []extractedJob) {
	c := make(chan extractedJob)
	var jobs []extractedJob
	pageURL := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	jobCards := doc.Find(".tapItem")
	jobCards.Each(func(i int, item *goquery.Selection) {
		go extractJob(item, c)
	})
	for i := 0; i < jobCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}
	mainC <- jobs

}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".jobTitle>span").Text())
	company := CleanString(card.Find(".companyName").Text())
	location := CleanString(card.Find(".companyLocation").Text())

	c <- extractedJob{
		id:       id,
		title:    title,
		company:  company,
		location: location,
	}
}

//CleanString cleans a string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}
