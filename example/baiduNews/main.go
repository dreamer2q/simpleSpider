package baiduNews

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"spider"
)

func main() {
	sp := &spider.Spider{
		URL:                spider.URL{"https://news.baidu.com/"},
		SpiderDepth:        1,
		Goroutines:         1,
		ActionTimeout:      0,
		AllowedDomain:      spider.Domain{"news.baidu.com"},
		AllowedContentType: spider.ContentType{spider.HtmlType},
		PreAction: func(req *http.Request) bool {
			fmt.Println("visit", req.URL)
			return false
		},
		PostAction: func(req *http.Request, resp *http.Response, body *bytes.Buffer) bool {

			doc, err := goquery.NewDocumentFromReader(body)
			if err != nil {
				log.Println("post", req.URL, err)
				return true
			}
			var category string
			doc.Find("div.menu-list").Find("li.active").EachWithBreak(func(i int, selection *goquery.Selection) bool {
				category = selection.Find("a").Text()
				fmt.Println("类别", category)
				return false //break loop
			})
			visitor := func(i int, selection *goquery.Selection) {
				fmt.Println(category, selection.AttrOr("href", ""), selection.Text())
			}
			doc.Find("a[target=\"_blank\"]").Each(visitor)
			return false
		},
		ExpandPolicy: func(req *http.Request, body *bytes.Buffer) spider.URL {
			expUrls := spider.URL{}
			doc, err := goquery.NewDocumentFromReader(body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error expand %s : %v\n", req.URL, err)
				return expUrls
			}
			doc.Find("div #channel-all").EachWithBreak(func(i int, selection *goquery.Selection) bool {
				selection.Find("a").Each(func(i int, selection *goquery.Selection) {
					if relateLink, b := selection.Attr("href"); b {
						if u, err := req.URL.Parse(relateLink); err == nil {
							expUrls.Add(u.String())
						}
					}
				})
				return false //block
			})
			return expUrls
		},
		OnError: func(req *http.Request, err error) {
			log.Println("error: visitting %s : %v", req.URL, err)
		},
	}
	log.Fatalln(sp.Run())
}
