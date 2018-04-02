package main

import (
	"bytes"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

var (
	linkstring = "http://shopping.friday.tw/shopping/Browse.do?op=vp&sid=12&cid=293368&pid=492408"
)

func TestSendLinkToChan(t *testing.T) {
	node := new(linkData)
	node.linkType = 1
	node.linkString = linkstring
	LinkChan <- node
	getnode := <-LinkChan
	if getnode.linkType == 1 {
		t.Log("pass")
	}
}

func TestParseHtmlStore1(t *testing.T) {
	body := []byte(`<div class="product_introduction">
		 <div class="promo-box">
		<span class="promoBtn"></span>
		</div>
		<h2 class="promotional">aaaaa</h2>
		<h3 class="trade_Name">test1</h3>
		<dl>
		<p class="prodIntroduction_spacer"></p>
		<dt>Introduction</dt>
		<dd class="introduction">asdf</dd>
		</dl>
		<p class="prodIntroduction_spacer"></p>
		<dl class="altrow" itemprop="offers" itemscope itemtype="http://schema.org/Offer" >

		<dt>original_price</dt>
		<dd class="original_price">125</dd>

		<dt>spe_price</dt>
		<dd class="list_price" itemprop="price">
				92 <span class="ntd">元</span>
		<span itemprop="priceCurrency" content="NTD"></span>
		<span itemprop="availability" content="http://schema.org/OnlineOnly"></span>
	    </dd>
		`)
	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(body))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".product_introduction").Each(func(i int, s *goquery.Selection) {
		productString := s.Find(".trade_Name").Text()
		price := s.Find(".list_price").Text()
		priceString := strings.Replace(price, "\n", "", -1)
		priceString = strings.Replace(priceString, "\t", "", -1)
		priceString = strings.Replace(priceString, " ", "", -1)
		re := regexp.MustCompile("[0-9]+")
		newPrice := re.FindAllString(priceString, -1)
		if intValue, err := strconv.Atoi(newPrice[0]); err == nil {
			if intValue == 92 {
				t.Log("pass")
			}
		}
		if productString == "test1" {
			t.Log("pass")
		}
	})
}
