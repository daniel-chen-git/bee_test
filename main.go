package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

type products struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Price      string `json:"price"`
	Category   string `json:"category"`
	Dimension2 string `json:"dimension2"`
	Dimension3 string `json:"dimension3"`
	Dimension4 string `json:"dimension4"`
}

type click struct {
	Products []products
}

type ecommerce struct {
	Click click
}

type production struct {
	Event     string `json:"event"`
	Ecommerce ecommerce
}

type linkData struct {
	linkType   int
	linkString string
}

var (
	StoreDomain1       = "shopping.friday.tw"
	StoreDomain2       = "www.rt-mart.com.tw"
	LinkChan           chan *linkData
	MaxQueryWorker     = 2
	ProductionDataMap1 = struct {
		sync.RWMutex
		procucts map[string]int
	}{procucts: make(map[string]int)}
	ProductionDataMap2 = struct {
		sync.RWMutex
		procucts map[string]int
	}{procucts: make(map[string]int)}
)

func init() {
	LinkChan = make(chan *linkData, 1000000)
}

func sendLinkToChan(linkType int, url string) {
	fmt.Printf("sendLinkToChan %s\n", url)
	node := new(linkData)
	node.linkType = linkType
	node.linkString = url
	fmt.Printf("sendLinkToChan %v\n", node)
	LinkChan <- node
}

func getDomainLink(domainstring string) {
	// Instantiate default collector
	var c *colly.Collector
	if domainstring == StoreDomain1 {
		c = colly.NewCollector(
			// Visit only domains: hackerspaces.org, wiki.hackerspaces.org
			colly.AllowedDomains(domainstring),
			colly.URLFilters(regexp.MustCompile("http://shopping\\.friday\\.tw/shopping/Browse\\.do\\?op\\=vp\\&sid\\=[0-9]*\\&cid\\=[0-9]*\\&pid\\=[0-9]*"), regexp.MustCompile("http://shopping\\.friday\\.tw/shopping/.+")),
			//colly.URLFilters(regexp.MustCompile("http://shopping\\.friday\\.tw/shopping/.+")),
		)
	} else if domainstring == StoreDomain2 {
		c = colly.NewCollector(
			colly.AllowedDomains("www.rt-mart.com.tw"))
	}

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		//fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL.String())
		if domainstring == StoreDomain1 {
			re := regexp.MustCompile("pid")
			findString := re.MatchString(r.URL.String())
			if findString {
				sendLinkToChan(1, r.URL.String())
			}
		} else if domainstring == StoreDomain2 {
			re := regexp.MustCompile("product_detail")
			findString := re.MatchString(r.URL.String())
			if findString {
				sendLinkToChan(2, r.URL.String())
			}
		}
	})

	// Start scraping on https://hackerspaces.org
	if domainstring == StoreDomain1 {
		c.Visit("http://shopping.friday.tw/shopping/1/s/12/")
	} else if domainstring == StoreDomain2 {
		c.Visit("http://www.rt-mart.com.tw/direct/")
	}
}

func parseHtmlStore1(node *linkData) {
	// Request the HTML page.
	//res, err := http.Get("http://shopping.friday.tw/shopping/Browse.do?op=vp&sid=12&cid=293368&pid=492408")
	res, err := http.Get(node.linkString)
	//res, err := http.Get("http://shopping.friday.tw/shopping/Browse.do?op=vc&cid=298540&sid=12")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".product_introduction").Each(func(i int, s *goquery.Selection) {
		//doc.Find(".product_Titlename , .price_num").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		productString := s.Find(".trade_Name").Text()
		price := s.Find(".list_price").Text()
		priceString := strings.Replace(price, "\n", "", -1)
		priceString = strings.Replace(priceString, "\t", "", -1)
		priceString = strings.Replace(priceString, " ", "", -1)
		//fmt.Printf("i = %d,s=%s\n", i, s.Text())
		//fmt.Printf("%d\n", len(s.Text()))
		//fmt.Printf("i=%d,a=%s\n", i, aa)
		//fmt.Printf("byte a=%v\n", []byte(aa))
		//fmt.Printf("len a=%d\n", len(aa))
		re := regexp.MustCompile("[0-9]+")
		newPrice := re.FindAllString(priceString, -1)
		fmt.Printf("==%s", newPrice[0])
		fmt.Printf("  ==%s\n", productString)
		if intValue, err := strconv.Atoi(newPrice[0]); err == nil {
			ProductionDataMap1.Lock()
			ProductionDataMap1.procucts[productString] = intValue
			ProductionDataMap1.Unlock()
		}
	})
}

func parseHtmlStore2(node *linkData) {
	// Request the HTML page.
	res, err := http.Get(node.linkString)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("script:contains(productClick)").Each(func(i int, s *goquery.Selection) {
		//doc.Find("script").Each(func(i int, s *goquery.Selection) {
		//fmt.Printf("i = %d,s=%s\n", i, s.Text())
		productString := strings.Replace(s.Text(), "\\", "", -1)
		productString = strings.Replace(productString, "dataLayer.push(", "", -1)
		productString = strings.Replace(productString, "\n", "", -1)
		productString = strings.Replace(productString, ");", "", -1)
		//fmt.Printf("i = %d,aa=%s\n", i, aa)
		productionJsonString := production{}
		if err := json.Unmarshal([]byte(productString), &productionJsonString); err != nil {
			panic(err)
		}
		fmt.Printf("  %v\n", productionJsonString.Ecommerce.Click.Products[0].Name)
		fmt.Printf("  %v\n", productionJsonString.Ecommerce.Click.Products[0].Price)
		//fmt.Printf("%v\n", len(e.Ecommerce.Click.Products))
		if intValue, err := strconv.Atoi(productionJsonString.Ecommerce.Click.Products[0].Price); err == nil {
			ProductionDataMap2.Lock()
			ProductionDataMap2.procucts[productionJsonString.Ecommerce.Click.Products[0].Name] = intValue
			ProductionDataMap2.Unlock()
		}
	})

}

func parseHtml(node *linkData) {
	if node.linkType == 1 {
		parseHtmlStore1(node)
	} else if node.linkType == 2 {
		parseHtmlStore2(node)
	}
}

func getLinkPage() {
	for {
		select {
		case node := <-LinkChan:
			parseHtml(node)
		default:
			time.Sleep(time.Second * 1)
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.GET("/query/", func(c *gin.Context) {
		storeName := c.DefaultQuery("name", "friday")
		product := c.DefaultQuery("productsName", "")
		re := regexp.MustCompile("^friday$")
		var ok bool
		var price int
		if re.MatchString(storeName) {
			ProductionDataMap1.RLock()
			price, ok = ProductionDataMap1.procucts[product]
			ProductionDataMap1.RUnlock()
		}
		re = regexp.MustCompile("^rt-mart$")
		if re.MatchString(storeName) {
			ProductionDataMap2.RLock()
			price, ok = ProductionDataMap2.procucts[product]
			ProductionDataMap2.RUnlock()
		}
		if ok {
			c.JSON(200, gin.H{"products": product, "price": price})
		} else {
			c.JSON(200, gin.H{"products": "Not find product"})
		}
	})
	r.GET("/Compare/", func(c *gin.Context) {
		product := c.DefaultQuery("productsName", "")
		ProductionDataMap1.RLock()
		price1, find1 := ProductionDataMap1.procucts[product]
		ProductionDataMap1.RUnlock()
		ProductionDataMap2.RLock()
		price2, find2 := ProductionDataMap2.procucts[product]
		ProductionDataMap2.RUnlock()
		if find1 && find2 {
			c.JSON(200, gin.H{"products": product, "price": price1, "price2": price2})
		} else {
			c.JSON(200, gin.H{"products": "Not find product"})
		}
	})
	return r
}

func main() {

	go getDomainLink(StoreDomain1)
	go getDomainLink(StoreDomain2)

	for i := 0; i < MaxQueryWorker; i++ {
		go getLinkPage()
	}

	r := setupRouter()
	r.Run(":8080")
}
