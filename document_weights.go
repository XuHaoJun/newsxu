package newsxu

import (
	"math"
)

type DocumentWeightsDumpDB struct {
	Id      string             `bson:"id" json:"id"`
	Weights map[string]float64 `bson:"weights" json:"weights"`
}

type DocumentWeights map[string]map[string]float64

type DocumentWeight struct {
	Id     string  `bson:"id" json:"id"`
	Weight float64 `bson:"weight" json:"weight"`
}

type FinalWeights []DocumentWeight

func (s FinalWeights) Len() int {
	return len(s)
}

func (s FinalWeights) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

func (s FinalWeights) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func NewDocumentWeightsByInvertedIndex(docs []Documenter, invertedIndex InvertedIndex) DocumentWeights {
	docWeights := make(map[string]map[string]float64, len(docs))
	docsLength := len(docs)
	for _, doc := range docs {
		segmentsLength := len(doc.Segments())
		termWeights := make(map[string]float64, segmentsLength)
		for _, s := range doc.Segments() {
			term := s.Token().Text()
			df := float64(len(invertedIndex[term]))
			idf := math.Log10(float64(docsLength) / df)
			nodes := invertedIndex[term]
			var tf float64
			for _, node := range nodes {
				if node.Id == doc.Id() {
					tf = float64(node.TermFrequency) / float64(segmentsLength)
					break
				}
			}
			termWeights[term] = tf * idf
		}
		docWeights[doc.Id()] = termWeights
	}
	return docWeights
}
