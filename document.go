package newsxu

import (
	"github.com/guotie/sego"
)

type Documenter interface {
	Id() string
	Text() string
	Segments() []sego.Segment
	SetSegments(ss []sego.Segment)
}

func NewDocumentsByNews(newss []*News) []Documenter {
	docs := make([]Documenter, len(newss))
	for i, news := range newss {
		docs[i] = NewNewsDocument(news)
	}
	return docs
}
