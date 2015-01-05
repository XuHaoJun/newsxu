package newsxu

type TermPosition struct {
	Start int `bson:"start" json:"start"`
	End   int `bson:"end" json:"end"`
}

type Node struct {
	Id            string
	TermFrequency int
	TermPositions []TermPosition
	Document      Documenter
}

type InvertedIndexNodeDumpDB struct {
	Id    string       `bson:"id" json:"id"`
	Nodes []NodeDumpDB `bson:"nodes" json:"nodes"`
}

type NodeDumpDB struct {
	DocumentId    string         `bson:"documentId" json:"documentId"`
	TermFrequency int            `bson:"termFrequency" json:"termFrequency"`
	TermPositions []TermPosition `bson:"termPositions" json:"termPositions"`
}

func (n *Node) DumpDB() NodeDumpDB {
	return NodeDumpDB{
		n.Id,
		n.TermFrequency,
		n.TermPositions,
	}
}

func (n *NodeDumpDB) Load() *Node {
	return &Node{
		n.DocumentId,
		n.TermFrequency,
		n.TermPositions,
		nil,
	}
}
