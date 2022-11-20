package newsxu

import (
	"os"

	mgo "github.com/qiniu/qmgo"
)

func LoadMongoConfig() *mgo.Config {
	return &mgo.Config{Uri: os.Getenv("MONGO_URL")}
}
