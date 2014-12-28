package newsxu

type Node struct {
	Id            string
	TermFrequency int
	Start         int
	End           int
	Document      Documenter
}

type InvertedIndexNodeDumpDB struct {
	Id    string       `bson:"id" json:"id"`
	Nodes []NodeDumpDB `bson:"nodes" json:"nodes"`
}

type NodeDumpDB struct {
	DocumentId    string `bson:"documentId" json:"documentId"`
	TermFrequency int    `bson:"termFrequency" json:"termFrequency"`
	Start         int    `bson:"start"  json:"start"`
	End           int    `bson:"end" json:"end"`
}

func (n *Node) DumpDB() NodeDumpDB {
	return NodeDumpDB{
		n.Id,
		n.TermFrequency,
		n.Start,
		n.End,
	}
}

func (n *NodeDumpDB) Load() *Node {
	return &Node{
		n.DocumentId,
		n.TermFrequency,
		n.Start,
		n.End,
		nil,
	}
}
