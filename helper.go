package spider

import (
	"bytes"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func appendSlice(m []string, check func(ele string) bool, arg ...string) []string {
	for _, v := range arg {
		if !check(v) {
			m = append(m, v)
		}
	}
	return m
}

func forEachNode(n *html.Node, pre, post func(n *html.Node) bool) {
	if n != nil {
		if pre != nil && pre(n) {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			forEachNode(c, pre, post)
		}
		if post != nil && post(n) {
			return
		}
	}
}

func getAttr(n []html.Attribute, key string) string {
	for _, v := range n {
		if v.Key == key {
			return v.Val
		}
	}
	return ""
}

func defaultExpandPolicy(req *http.Request, body *bytes.Buffer) URL {
	doc, err := html.Parse(body)
	if err != nil {
		log.Printf("defaultExpandPolicy: parsing %s as HTML %v\n", req.URL.String(), err)
		return nil
	}
	retUrl := make(URL, 0)
	forEachNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			var t string
			switch n.Data {
			case "a", "link":
				t = getAttr(n.Attr, "href")
			case "img", "script":
				t = getAttr(n.Attr, "src")
			case "style":
				if n.FirstChild != nil && n.FirstChild.Data != "" {
					reg := regexp.MustCompile("url\\((.+?)\\)")
					urls := reg.FindAllStringSubmatch(n.FirstChild.Data, -1)
					for _, u := range urls {
						if u1, err := req.URL.Parse(u[1]); err == nil {
							retUrl = append(retUrl, u1.String())
						}
					}
				}
			}
			if t != "" {
				t2, err := req.URL.Parse(t)
				//fmt.Println("expand", t2, err)
				if err == nil && !strings.HasPrefix(t2.String(), "javascript") {
					retUrl = append(retUrl, t2.String())
				}
			}
		}
		return false
	}, nil)
	return retUrl
}

func (s *Spider) checkDomain(dm string) bool {
	//dm is hostname
	if dm == "" {
		//true means block
		return true
	}
	host, err := url.Parse(dm)
	if err != nil {
		log.Printf("checking domain %s\n", dm)
		return true
	}
	for _, v := range s.AllowedDomain {
		if strings.HasSuffix(host.Host, v) {
			//false means allow
			return false
		}
	}
	return true
}

func (s *Spider) checkContentType(ct string) bool {
	if ct == "" {
		return true
	}
	for _, v := range s.AllowedContentType {
		if b, _ := regexp.MatchString(v, ct); b {
			return false
		}
	}
	return true
}

func (u *URL) Add(urls ...string) {
	for _, l := range urls {
		*u = append(*u, l)
	}
}
