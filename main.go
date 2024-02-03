package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
)

type jumiaItems struct {
	Name          string
	DiscountPrice string
	OldPrice      string
	Discount      string
	Image         string
	Rating        string
	Merchant      string
}

func (j *jumiaItems) addToCsv(csvWriter *csv.Writer) {
	row := []string{j.Name, j.DiscountPrice, j.OldPrice, j.Discount, j.Image, j.Rating, j.Merchant}
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

	file, err := os.Create("jumiaItems.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	csvWriter := initCsvWriter(file)
	defer csvWriter.Flush()

	c := colly.NewCollector()

	c.OnHTML(".prd._fb", func(h *colly.HTMLElement) {
		item := jumiaItems{}
		item.Image = checkElementAttr(h, ".img", "data-src")
		item.Merchant = checkElementAttr(h, ".xprss", "aria-label")

		item.Name = checkElementText(h, ".name")
		item.DiscountPrice = checkElementText(h, ".prc")
		item.OldPrice = checkElementText(h, ".old")
		item.Discount = checkElementText(h, "._dsct")
		item.Rating = checkElementText(h, ".rev")

		item.addToCsv(csvWriter)

	})

	c.OnHTML(".pg-w a", func(h *colly.HTMLElement) {
		if h.Attr("aria-label") == "Next Page" {
			nextPage := h.Request.AbsoluteURL(h.Attr("href"))
			fmt.Println(nextPage)
			c.Visit(nextPage)
		}
	})

	c.Visit("https://www.jumia.com.ng/mlp-fashion-deals/mens-fashion-accessories/")

	fmt.Printf("Process took %s",time.Since(startTime))
}

func initCsvWriter(file *os.File) *csv.Writer {
	writer := csv.NewWriter(file)

	headers := []string{"name", "discountPrice", "oldPrice", "discount", "image", "rating", "merchant"}
	writer.Write(headers)

	return writer
}
