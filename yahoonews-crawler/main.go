package main

import (
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	gourl "net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type YahooNews struct {
	Id       string `bson:"id"`
	Title    string `bson:"title"`
	Content  string `bson:"content"`
	Author   string `bson:"author"`
	Provider string `bson:"provider"`
	PostTime string `bson:"postTime"`
	URL      string `bson:"url" json:"url"`
}

func yahooNewURLs(url string) []string {
	doc, err := goquery.NewDocument(url)
	if err != nil {
    log.Println(err)
		return []string{}
	}
	result := doc.Find(".yom-list-wide li").Map(func(i int, s *goquery.Selection) string {
		newsLink, _ := s.Find(".story>.txt>h4>a").Attr("href")
		return "https://tw.news.yahoo.com" + newsLink
	})
	return result
}

func parseYahooNewURL(url string) (string, string) {
	u, _ := gourl.Parse(url)
	uri := strings.Trim(u.Path, "/")
	s := strings.Split(uri, "-")
	id := strings.Split(s[len(s)-1], ".html")[0]
	_, err := strconv.Atoi(id)
	idPosition := 1
	if err != nil {
		if s[len(s)-2] == "" {
			id = strings.Split(s[len(s)-3]+"-"+s[len(s)-1], ".html")[0]
			idPosition = 3
		}
	}
	title := strings.Join(s[:len(s)-idPosition], "")
	if title == "" {
		//log.Println("warnning zero title", url)
	}
	return id, title
}

func YahooNewURLs() []string {
	result := []string{}
	resultChan := make(chan []string, 40)
	for i := 1; i <= 40; i++ {
		go func(index int) {
			time.Sleep(time.Duration(i) * time.Millisecond * 100)
			pageURL := "https://tw.news.yahoo.com/archive/" + strconv.Itoa(index) + ".html"
			resultChan <- yahooNewURLs(pageURL)
		}(i)
	}
	for i := 1; i <= 40; i++ {
		subResult := <-resultChan
		result = append(result, subResult...)
	}
	return result
}

func yahooNewContent(url string) *YahooNews {
	doc, err := goquery.NewDocument(url)
	id, title := parseYahooNewURL(url)
	if err != nil {
		log.Println(err)
		return &YahooNews{
			Id:    id,
			Title: title,
		}
	}

	content := doc.Find(".yom-art-content").Find("p").Text()
	postTime := doc.Find(".byline.vcard>abbr").Text()
	provider := doc.Find(".byline.vcard>.provider").Text()
	author := doc.Find(".byline.vcard>.fn").Text()
	if content == "" {
		log.Println("warnning zero content", url)
	}
	return &YahooNews{
		Id:       id,
		Title:    title,
		Content:  content,
		Author:   author,
		Provider: provider,
		PostTime: postTime,
		URL:      url,
	}
}

func downloadYahooNews() {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// 下載新聞
	log.Println("開始下載新聞 每次連接間隔時間： 0.1s")
	startTime := time.Now()
	urls := YahooNewURLs()
	newsChan := make(chan *YahooNews, len(urls))
	for index, url := range urls {
		go func(i int, u string) {
			time.Sleep(time.Duration(i) * time.Millisecond * 100)
			newsChan <- yahooNewContent(u)
		}(index, url)
	}

	c := session.DB("sego").C("yahooNews")

	for i := 0; i < len(urls); i++ {
		ynews := <-newsChan
		log.Println("下載數量:", i+1, "標題：", ynews.Title, "Id: ", ynews.Id)
		if ynews != nil && ynews.Title != "" && ynews.Id != "" && ynews.Title != " " && ynews.Content != "" {
			c.Upsert(bson.M{"id": ynews.Id}, ynews)
		}
	}

	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
	log.Println("新聞下載完畢", "共耗時:", time.Since(startTime))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	downloadYahooNews()
	tick := time.Tick(2 * time.Hour)
	for range tick {
		go downloadYahooNews()
	}
}
