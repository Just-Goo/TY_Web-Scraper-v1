package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

var jumia = []string{
	"https://www.jumia.com.ng/electronic-accessories-supplies/",
	"https://www.jumia.com.ng/camera-photo-accessories/",
	"https://www.jumia.com.ng/electronic-supplies-power-protection/",
	"https://www.jumia.com.ng/microphones/",
	"https://www.jumia.com.ng/electronics-cables/",
}

type jumiaItems struct {
	Name          string
	DiscountPrice string
	OldPrice      string
	Discount      string
	Image         string
	Rating        string
	Category      string
	Merchant      string
}

func (j *jumiaItems) addToCsv(csvWriter *csv.Writer) {
	row := []string{j.Name, j.DiscountPrice, j.OldPrice, j.Discount, j.Image, j.Rating, j.Category, j.Merchant}
	csvWriter.Write(row)
}

func checkElementText(h *colly.HTMLElement, query string) string {
	if value := h.ChildText(query); value != "" {
		return value
	}
	return "not available"
}

func checkElementAttr(h *colly.HTMLElement, query, attrName string) string {
	if value := h.ChildAttr(query, attrName); value != "" {
		return value
	}
	return "not available"
}

func main() {

	startTime := time.Now()

	var notifyDoneNum = 0

	itemChan := make(chan jumiaItems)
	notifyDoneChan := make(chan int)

	file, err := os.Create("jumiaItems.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	csvWriter := initCsvWriter(file)
	defer csvWriter.Flush()

	for index, category := range jumia {
		go scrapeJumiaCategories(category, index, &notifyDoneNum, itemChan, notifyDoneChan)
	}

CHAN:
	for {
		select {
		case item := <-itemChan:
			item.addToCsv(csvWriter)
		case notifyDoneNum = <-notifyDoneChan:   
			if notifyDoneNum == 4 {
				break CHAN
			}
		}
	}

	fmt.Printf("Process took %s", time.Since(startTime))
}

func getCategoryFromIndex(index int) string {
	if index == 0 {
		return "electronic accessories supplies"
	}
	if index == 1 {
		return "camera photo accessories"
	}
	if index == 2 {
		return "electronic supplies power protection"
	}
	if index == 3 {
		return "microphones"
	}
	if index == 4 {
		return "electronic cables"
	}
	return "another category"
}

func scrapeJumiaCategories(categoryLink string, index int, notifyDoneNum *int, itemChan chan<- jumiaItems, notifyDoneChan chan<- int) {

	c := colly.NewCollector()
	var lastPage int
	var pageIndex string

	c.OnHTML(".prd._fb", func(h *colly.HTMLElement) {
		item := jumiaItems{}
		item.Image = checkElementAttr(h, ".img", "data-src")
		item.Merchant = checkElementAttr(h, ".xprss", "aria-label")

		item.Name = checkElementText(h, ".name")
		item.DiscountPrice = checkElementText(h, ".prc")
		item.OldPrice = checkElementText(h, ".old")
		item.Discount = checkElementText(h, "._dsct")
		item.Rating = checkElementText(h, ".rev")
		item.Category = getCategoryFromIndex(index)

		itemChan <- item
	})

	c.OnHTML(".pg-w a", func(h *colly.HTMLElement) {
		// Get the last page
		if lastPage == 0 {
			if h.Attr("aria-label") == "Last Page" {
				nextPage := h.Request.AbsoluteURL(h.Attr("href"))
				pg := nextPage[len(nextPage)-2:]
				if strings.HasPrefix(pg, "=") {
					pg = pg[len(pg)-1:]
				}

				lp, err := strconv.Atoi(pg)

				if err != nil {
					log.Fatal(err)
				}

				
				if pageIndex == strconv.Itoa(lp) {
					notifyDoneChan <- *notifyDoneNum + 1
				} 

				lastPage = lp
			}
		}

		if h.Attr("aria-label") == "Next Page" {
			nextPage := h.Request.AbsoluteURL(h.Attr("href"))
			pg := nextPage[len(nextPage)-2:]
			if strings.HasPrefix(pg, "=") {
				pg = pg[len(pg)-1:]
			}
			pageIndex = pg

			fmt.Println(nextPage)

			c.Visit(nextPage)
		}
	})

	c.Visit(categoryLink)

}

func initCsvWriter(file *os.File) *csv.Writer {
	writer := csv.NewWriter(file)

	headers := []string{"name", "discountPrice", "oldPrice", "discount", "image", "rating", "category", "merchant"}
	writer.Write(headers)

	return writer
}
