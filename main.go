package main

import (
	"os"
	"strings"

	"github.com/GoJobScrapper/scrapper"
	"github.com/labstack/echo/v4"
)

const fileName = "jobs.csv"

func handleScrape(c echo.Context) error {
	defer os.Remove("jobs.csv")
	keyword := strings.ToLower(scrapper.CleanString(c.FormValue("keyword")))
	scrapper.Scrapper(keyword)
	return c.Attachment("jobs.csv", "jobs.csv")
}

func handleHome(c echo.Context) error {

	return c.File("home.html")
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/results", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}
