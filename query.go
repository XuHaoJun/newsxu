package newsxu

import (
	"github.com/guotie/sego"
)

type QueryDocument struct {
	id       string
	segments []sego.Segment
	text     string
}

func NewQueryDocument(id string, text string) *QueryDocument {
	return &QueryDocument{
		id:   id,
		text: text,
	}
}

func (d *QueryDocument) Text() string {
	return d.text
}

func (d *QueryDocument) Segments() []sego.Segment {
	return d.segments
}

func (d *QueryDocument) SetSegments(ss []sego.Segment) {
	d.segments = ss
}

func (d *QueryDocument) Id() string {
	return d.id
}
