package newsxu

import (
	"math"

	"github.com/guotie/sego"
)

type QueryWeights map[string]float64
type QueryNodes map[string][]*Node

func NewQueryWeights(queryDocument *QueryDocument, docInvertedIndex InvertedIndexer, segmenter *sego.Segmenter, stopword *sego.StopWords) (QueryWeights, QueryNodes) {
	queryInvertedIndex := NewInvertedIndexBySego([]Documenter{queryDocument}, segmenter, stopword)
	querySegments := queryDocument.Segments()
	queryWeights := make(map[string]float64, len(querySegments))
	queryNodes := make(map[string][]*Node, len(querySegments))
	foundCount := 0
	for _, s := range querySegments {
		term := s.Token().Text()
		var weight float64 = 0
		nodes := docInvertedIndex.Find(term)
		if nodes != nil {
			queryNodes[term] = nodes
			tf := float64(queryInvertedIndex[term][0].TermFrequency)
			// weight = tf / math.Sqrt(float64(len(querySegments)))
			weight = tf
			foundCount++
		}
		queryWeights[term] = weight
	}
	for term, weight := range queryWeights {
		queryWeights[term] = weight / math.Sqrt(float64(foundCount))
	}
	return queryWeights, queryNodes
}
