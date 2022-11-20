package main

import (
	"context"

	"github.com/guotie/sego"
	mgo "github.com/qiniu/qmgo"
	opts "github.com/qiniu/qmgo/options"
	"github.com/samber/lo"
	"github.com/xuhaojun/newsxu"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"log"
	"runtime"
	"sync"
	"time"
)

func getYahooNewss() []*newsxu.News {
	session, err := mgo.NewClient(context.Background(), newsxu.LoadMongoConfig())
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer session.Close(context.Background())
	c := session.Database("sego").Collection("yahooNews")

	var ynewss []*newsxu.News
	c.Find(context.Background(), bson.M{}).All(&ynewss)
	return ynewss
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Println("載入新聞")
	docs := newsxu.NewDocumentsByNews(getYahooNewss())
	log.Println("新聞向量長度：", len(docs))

	segmenter := &sego.Segmenter{}
	stopword := &sego.StopWords{}
	segmenter.LoadDictionary("../data/dictionary.txt")
	stopword.LoadDictionary("../data/stopwords-utf8.txt")
	startTime := time.Now()
	log.Println("建立 inverted index")
	invertedIndex := newsxu.NewInvertedIndexBySego(docs, segmenter, stopword)
	// for k, v := range invertedIndex {
	// 	fmt.Print(k)
	// 	fmt.Print(" [ ")
	// 	for _, node := range v {
	// 		fmt.Print("(" + node.Id + " ")
	// 		fmt.Print(strconv.Itoa(node.TermFrequency) + ") ")
	// 	}
	// 	fmt.Print("]\n")
	// }
	log.Println("完成 inverted index 長度：", len(invertedIndex), "共耗時：", time.Since(startTime))

	startTime = time.Now()
	log.Println("建立 weight 表")
	docWeights := newsxu.NewDocumentWeightsByInvertedIndex(docs, invertedIndex)
	log.Println("完成 weight 表, 共耗時： ", time.Since(startTime))
	// for k, v := range docWeights {
	// 	fmt.Print(k)
	// 	fmt.Print(" [ ")
	// 	for term, weight := range v {
	// 		fmt.Print("(" + term + " ")
	// 		fmt.Print(strconv.FormatFloat(weight, 'f', 6, 64) + ") ")
	// 	}
	// 	fmt.Print("]\n")
	// }

	startTime = time.Now()
	log.Println("更新資料庫和建立索引")
	session, err := mgo.NewClient(context.Background(), newsxu.LoadMongoConfig())
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer session.Close(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		invertC := session.Database("sego").Collection("invertedIndex")
		invertC.DropCollection(context.Background())
		bulk := invertC.Bulk()
		bulk.SetOrdered(false)
		count := 0
		for term, nodes := range invertedIndex {
			nodeDumpDBs := make([]newsxu.NodeDumpDB, len(nodes))
			for i, node := range nodes {
				nodeDumpDBs[i] = node.DumpDB()
			}
			bulk.InsertOne(newsxu.InvertedIndexNodeDumpDB{term, nodeDumpDBs})
			count++
			if count%1000 == 0 {
				_, err := bulk.Run(context.Background())
				if err != nil {
					log.Println(err)
				}
				bulk = invertC.Bulk()
				bulk.SetOrdered(false)
				count = 0
			}
			//invertC.Upsert(bson.M{"id": term},
			//  newsxu.InvertedIndexNodeDumpDB{term, nodeDumpDBs})
		}
		if count > 0 {
			_, err := bulk.Run(context.Background())
			if err != nil {
				log.Println(err)
			}
		}
		indexes := []opts.IndexModel{{
			Key: []string{"id"},
			IndexOptions: &options.IndexOptions{
				Unique: lo.ToPtr(true),
			},
		}}
		invertC.CreateIndexes(context.Background(), indexes)
		wg.Done()
	}()

	go func() {
		weightC := session.Database("sego").Collection("documentWeights")
		//for docId, weights := range docWeights {
		//	weightC.Upsert(bson.M{"id": docId},
		//		newsxu.DocumentWeightsDumpDB{docId, weights})
		//}
		weightC.DropCollection(context.Background())
		// bulk
		bulk := weightC.Bulk()
		bulk.SetOrdered(false)
		count := 0
		for docId, weights := range docWeights {
			bulk.InsertOne(newsxu.DocumentWeightsDumpDB{docId, weights})
			count++
			if count%1000 == 0 {
				_, err := bulk.Run(context.Background())
				if err != nil {
					log.Println(err)
				}
				bulk = weightC.Bulk()
				bulk.SetOrdered(false)
				count = 0
			}
		}
		if count > 0 {
			_, err := bulk.Run(context.Background())
			if err != nil {
				log.Println(err)
			}
		}
		// end of bulk
		indexes := []opts.IndexModel{{
			Key: []string{"id"},
			IndexOptions: &options.IndexOptions{
				Unique: lo.ToPtr(true),
			},
		}}
		weightC.CreateIndexes(context.Background(), indexes)
		wg.Done()
	}()
	wg.Wait()
	log.Println("完成更新資料庫, 共耗時： ", time.Since(startTime))

	// startTime = time.Now()
	// queryText := "我家的貓咪圓又圓飛天跳躍神奇小花貓貓咪跳跳唷！"
	// queryDocument := newsxu.NewQueryDocument("query", queryText)
	// queryWeights, _ := newsxu.NewQueryWeights(queryDocument, invertedIndex, segmenter, stopword)
	// log.Println(queryWeights)
	// log.Println("完成查詢, 共耗時： ", time.Since(startTime))
}
