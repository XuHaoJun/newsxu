package newsxu

import (
	"github.com/guotie/sego"
)

type News struct {
	Id       string `bson:"id" json:"id"`
	Title    string `bson:"title" json:"title"`
	Content  string `bson:"content" json:"content"`
	Author   string `bson:"author" json:"author"`
	Provider string `bson:"provider" json:"provider"`
	PostTime string `bson:"postTime" json:"postTime"`
	URL      string `bson:"url" json:"url"`
}

type NewsDocument struct {
	segments []sego.Segment
	news     *News
}

func NewNewsDocument(news *News) *NewsDocument {
	return &NewsDocument{nil, news}
}

func (d *NewsDocument) Id() string {
	return d.news.Id
}

func (d *NewsDocument) Text() string {
	return d.news.Content
}

func (d *NewsDocument) Segments() []sego.Segment {
	return d.segments
}

func (d *NewsDocument) SetSegments(ss []sego.Segment) {
	d.segments = ss
}
