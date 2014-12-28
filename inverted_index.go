package newsxu

import (
	"github.com/guotie/sego"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
)

type InvertedIndex map[string][]*Node

type InvertedIndexDB struct {
	Key string
	C   *mgo.Collection
}

func (invDB *InvertedIndexDB) Find(term string) []*Node {
	c := invDB.C
	dump := &InvertedIndexNodeDumpDB{}
	err := c.Find(bson.M{invDB.Key: term}).One(dump)
	if err != nil {
		return nil
	}
	nodes := make([]*Node, len(dump.Nodes))
	for i, nodeDump := range dump.Nodes {
		nodes[i] = nodeDump.Load()
	}
	return nodes
}

func (inv InvertedIndex) Find(term string) []*Node {
	return inv[term]
}

type InvertedIndexer interface {
	Find(term string) []*Node
}

// TODO
// add term's pos return!

func NewInvertedIndexBySego(docs []Documenter, segmenter *sego.Segmenter, stopword *sego.StopWords) InvertedIndex {
	invertedIndex := make(map[string][]*Node)
	wg := &sync.WaitGroup{}
	wg.Add(len(docs))
	for _, d := range docs {
		go func(doc Documenter) {
			segments := segmenter.Segment([]byte(doc.Text()))
			filtedSegments := stopword.Filter(segments, true)
			doc.SetSegments(filtedSegments)
			wg.Done()
		}(d)
	}
	wg.Wait()
	for _, doc := range docs {
		for _, s := range doc.Segments() {
			token := s.Token()
			term := token.Text()
			list := invertedIndex[term]
			if list == nil {
				list = []*Node{&Node{doc.Id(), 1, s.Start(), s.End(), doc}}
			} else {
				isDupNode := false
				for _, node := range list {
					if node.Id == doc.Id() {
						node.TermFrequency += 1
						isDupNode = true
						break
					}
				}
				if !isDupNode {
					list = append(list, &Node{doc.Id(), 1, s.Start(), s.End(), doc})
				}
			}
			invertedIndex[term] = list
		}
	}
	return invertedIndex
}
