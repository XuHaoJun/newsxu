package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"sort"
	"time"

	"github.com/guotie/sego"
	mgo "github.com/qiniu/qmgo"
	"github.com/samber/lo"
	"github.com/xuhaojun/newsxu"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	host         = flag.String("host", "", "HTTP服务器主机名")
	port         = flag.Int("port", 8080, "HTTP服务器端口")
	dict         = flag.String("dict", "../data/dict.txt.small", "词典文件")
	stopwordTxt  = flag.String("stopword", "../data/stopwords-utf8.txt", "")
	staticFolder = flag.String("static_folder", "public", "静态页面存放的目录")
	segmenter    = &sego.Segmenter{}
	stopword     = &sego.StopWords{}
	dbSession    = &mgo.Client{}
)

type News struct {
	Id       string  `bson:"id" json:"id"`
	Weight   float64 `bson:",omitempty" json:"weight"`
	Title    string  `bson:"title" json:"title"`
	URL      string  `bson:"url" json:"url"`
	Provider string  `bson:"provider" json:"provider"`
	PostTime string  `bson:"postTime" json:"postTime"`
}

type JsonResponse struct {
	QueryWeights   newsxu.QueryWeights `json:"queryWeights"`
	Newss          []*News             `json:"newss"`
	SearchUsedTime string              `json:"searchUsedTime"`
}

func JsonRpcServer(w http.ResponseWriter, req *http.Request) {
	//
	session := dbSession
	db := session.Database("sego")
	yahooNews := db.Collection("yahooNews")
	invertedIndex := &newsxu.InvertedIndexDB{
		Key: "id",
		C:   db.Collection("invertedIndex"),
	}
	documentWeights := db.Collection("documentWeights")

	//得到要分词的文本
	queryText := req.URL.Query().Get("text")
	if queryText == "" {
		queryText = req.PostFormValue("text")
	}

	startTime := time.Now()
	// log.Println("接受查詢 text:", queryText)
	queryDocument := newsxu.NewQueryDocument("query", queryText)
	queryWeights, queryNodes := newsxu.NewQueryWeights(queryDocument, invertedIndex, segmenter, stopword)
	docIds := make([]string, 0)
	for _, nodes := range queryNodes {
		for _, node := range nodes {
			found := false
			for _, docId := range docIds {
				if docId == node.Id {
					found = true
					break
				}
			}
			if !found {
				docIds = append(docIds, node.Id)
			}
		}
	}
	docWeights := make(map[string]map[string]float64, len(docIds))
	documentWeightDumps := []*newsxu.DocumentWeightsDumpDB{}
	documentWeights.Find(context.Background(), bson.M{"id": bson.M{"$in": docIds}}).All(&documentWeightDumps)
	for _, dump := range documentWeightDumps {
		docWeights[dump.Id] = dump.Weights
	}

	finalWeights := make(map[string]float64, len(docIds))
	for docId, docWeight := range docWeights {
		finalWeights[docId] = 0
		for queryTerm, qw := range queryWeights {
			dw, ok := docWeight[queryTerm]
			if ok {
				finalWeights[docId] += dw * qw
			}
		}
	}
	finalWeightsSlice := make(newsxu.FinalWeights, len(finalWeights))
	i := 0
	for docId, docWeight := range finalWeights {
		finalWeightsSlice[i] = newsxu.DocumentWeight{docId, docWeight}
		i++
	}
	sort.Sort(finalWeightsSlice)

	// log.Println("query weights:", queryWeights)
	// log.Println("query docIds:", docIds)
	// log.Println("doc Weights", docWeights)
	// log.Println("final Weights", finalWeightsSlice)

	newss := []*News{}
	yahooNews.Find(context.Background(),
		bson.M{"id": bson.M{"$in": lo.Map(finalWeightsSlice, func(item newsxu.DocumentWeight, index int) string {
			return item.Id
		})}}).
		Select(bson.M{"title": 1, "provider": 1, "url": 1, "postTime": 1, "id": 1}).
		All(&newss)
	lo.ForEach(newss, func(newsItem *News, index int) {
		weight, found := lo.Find(finalWeightsSlice, func(weightItem newsxu.DocumentWeight) bool {
			return newsItem.Id == weightItem.Id
		})
		if found {
			newsItem.Weight = weight.Weight
		}
	})

	response, _ := json.Marshal(
		&JsonResponse{queryWeights, newss,
			time.Since(startTime).String()})

	// log.Println("完成查詢, 共耗時： ", time.Since(startTime))

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(response))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	var err error
	dbSession, err = mgo.NewClient(context.Background(), newsxu.LoadMongoConfig())
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer dbSession.Close(context.Background())
	segmenter.LoadDictionary(*dict)
	stopword.LoadDictionary(*stopwordTxt)

	http.HandleFunc("/json", JsonRpcServer)
	http.Handle("/", http.FileServer(http.Dir(*staticFolder)))

	log.Print("服务器启动")
	http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
}
