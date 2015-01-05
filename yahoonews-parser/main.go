package main

import (
	"github.com/guotie/sego"
	"github.com/xuhaojun/newsxu"
	"gopkg.in/mgo.v2"
	//"gopkg.in/mgo.v2/bson"
	"log"
	"runtime"
	"sync"
	"time"
)

func getYahooNewss() []*newsxu.News {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer session.Close()
	c := session.DB("sego").C("yahooNews")

	var ynewss []*newsxu.News
	c.Find(nil).All(&ynewss)
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
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		session2 := session.Clone()
		invertC := session2.DB("sego").C("invertedIndex")
		invertC.DropCollection()
		bulk := invertC.Bulk()
		bulk.Unordered()
		count := 0
		for term, nodes := range invertedIndex {
			nodeDumpDBs := make([]newsxu.NodeDumpDB, len(nodes))
			for i, node := range nodes {
				nodeDumpDBs[i] = node.DumpDB()
			}
			bulk.Insert(newsxu.InvertedIndexNodeDumpDB{term, nodeDumpDBs})
			count++
			if count%1000 == 0 {
				_, err := bulk.Run()
				if err != nil {
					log.Println(err)
				}
				bulk = invertC.Bulk()
				bulk.Unordered()
				count = 0
			}
			//invertC.Upsert(bson.M{"id": term},
			//  newsxu.InvertedIndexNodeDumpDB{term, nodeDumpDBs})
		}
		if count > 0 {
			_, err := bulk.Run()
			if err != nil {
				log.Println(err)
			}
		}
		index := mgo.Index{
			Key:        []string{"id"},
			Unique:     true,
			DropDups:   true,
			Background: false,
			Sparse:     true,
		}
		invertC.EnsureIndex(index)
		session2.Close()
		wg.Done()
	}()

	go func() {
		session2 := session.Clone()
		weightC := session2.DB("sego").C("documentWeights")
		//for docId, weights := range docWeights {
		//	weightC.Upsert(bson.M{"id": docId},
		//		newsxu.DocumentWeightsDumpDB{docId, weights})
		//}
		weightC.DropCollection()
		// bulk
		bulk := weightC.Bulk()
		bulk.Unordered()
		count := 0
		for docId, weights := range docWeights {
			bulk.Insert(newsxu.DocumentWeightsDumpDB{docId, weights})
			count++
			if count%1000 == 0 {
				_, err := bulk.Run()
				if err != nil {
					log.Println(err)
				}
				bulk = weightC.Bulk()
				bulk.Unordered()
				count = 0
			}
		}
		if count > 0 {
			_, err := bulk.Run()
			if err != nil {
				log.Println(err)
			}
		}
		// end of bulk
		index := mgo.Index{
			Key:        []string{"id"},
			Unique:     true,
			DropDups:   true,
			Background: false,
			Sparse:     true,
		}
		weightC.EnsureIndex(index)
		session2.Close()
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
