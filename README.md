# Introductions:
 ## Crawler for 2 online stores and use restful API to query and compare


# How to use
 ## 1. Install golang
  ### 1.1 [GO](https://golang.org/doc/install)

 ## 2. Install packet
  ### go get -u github.com/PuerkitoBio/goquery/...
  ### go get -u github.com/gin-gonic/gin/...
  ### go get -u github.com/gocolly/colly/...

 ## 3. Build source code
  ### go build

 ## 4. exec
  ### ./bee_test

 ## 5. API
  ### 5.1 Query
   #### /query/?name=XXX&productsName=XXX
    ##### name only rt-mart or friday
    ##### productsName is product name
    ##### ex:localhost:8080/query/?name=rt-mart&productsName=RT Water
  ### 5.2 Compare
   #### /compare/?productsName=
    ##### productsName is product name
    ##### ex:localhost:8080/query/?productsName=RT Water

 ## 6. Port
  ### 6.1 8080
