package crontab

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/topics/crawler"
	"github.com/topics/models"
	"github.com/topics/sysexec"
)

type TAIEX struct {
	BasicCron
}

var TAIEXModel = new(models.TAIEXModel)

func (m *TAIEX) Period() string {
	// return "@hourly"
	return "0 * * * *"
}

func (m *TAIEX) Do() {
	// Check if there is a instance running, kill it
	if pid := sysexec.FindWebDriverPID(os.Getenv("CRAWLER_TAIEX_PORT")); pid != nil {
		sysexec.KillWebDriver(pid)
	}

	// Initialize
	crawlerEntry := crawler.CrawlerEntry{
		URL:             "https://www.twse.com.tw/zh/page/trading/indices/MI_5MINS_HIST.html",
		SeleniumPath:    os.Getenv("SELENIUM"),
		GeckoDriverPath: os.Getenv("GECKO_DRIVER"),
	}

	port, err := strconv.Atoi(os.Getenv("CRAWLER_TAIEX_PORT"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "WebDriver can't get correct port number"))
	}
	crawlerEntry.Port = port

	// Driver instance startup
	webDriver, err := crawlerEntry.StartWebInstance()
	if err != nil {
		log.Fatal(errors.Wrap(err, "WebDriver instance startup fail"))
	}
	// New crawler with URL
	crawlerEntry.Crawler, err = crawlerEntry.Init()
	if err != nil {
		log.Fatal(errors.Wrap(err, "URL connection fail"))
	}
	// Find the latest record in database, return 1970-01-01 if empty
	date := TAIEXModel.LatestDate()
	log.Printf("The latest date of TAIEX is %s", date)
	// Startup crawler with date(first day of current month)
	TAIEX, err := crawlerEntry.TAIEX(time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Get TAIEX fail"))
	}
	TAIEXModel.Store(TAIEX)
	// Stop crawler and web driver
	defer (*crawlerEntry.Crawler).Quit()
	defer webDriver.Stop()
}