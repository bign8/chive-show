package graph

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

// TODO: add some graph processing functions

// Graph is the serializable graph we have all been looking for
type Graph struct {
	s     *SerialGraph
	nodes map[uint64]*Node              // Optimal lookup with pointers goes here
	dupes map[NodeType]map[string]*Node // type > value > node
	edges map[uint64]map[uint64]bool    // Edge duplicate detection
}

// New creates a new Graph
func New(isDirected bool) *Graph {
	return &Graph{
		s: &SerialGraph{
			Nodes:     make([]*Node, 0),
			Directed:  proto.Bool(isDirected),
			NodeCount: proto.Uint64(0),
		},
		nodes: make(map[uint64]*Node),
		dupes: make(map[NodeType]map[string]*Node),
		edges: make(map[uint64]map[uint64]bool),
	}
}

// Add creates and adds a node to the graph
func (g *Graph) Add(value string, ttype NodeType, weight int64) *Node {

	// Check duplicate node (add weight)
	dupe := g.dupes[ttype][value]
	if dupe != nil {
		*dupe.Weight += weight
		return dupe
	}

	// Create new node
	n := &Node{
		Id:       proto.Uint64(g.genNodeID()),
		Value:    proto.String(value),
		Weight:   proto.Int64(weight),
		Type:     ttype.Enum(),
		Adjacent: make([]uint64, 0),
	}
	g.nodes[*n.Id] = n
	g.s.Nodes = append(g.s.Nodes, n)

	// Add dupe check to list
	dub, ok := g.dupes[ttype]
	if !ok {
		dub = make(map[string]*Node)
		g.dupes[ttype] = dub
	}
	dub[value] = n
	return n
}

// Connect connects nodes to and from with an edge of weight w
func (g *Graph) Connect(from, to *Node, weight int64) error {
	if to == nil || from == nil {
		return errors.New("Cannot add edge to nil node")
	}

	mm := g.edges[*from.Id]
	if mm == nil {
		mm = make(map[uint64]bool)
		g.edges[*from.Id] = mm
	}
	if !mm[*to.Id] {
		from.Adjacent = append(from.Adjacent, *to.Id) // Directed edge
		from.Weights = append(from.Weights, weight)
		mm[*to.Id] = true
	}

	if !g.s.GetDirected() && !g.edges[*to.Id][*from.Id] { // UnDirected edge (return trip)
		g.Connect(to, from, weight)
	}
	return nil
}

func (g *Graph) genNodeID() (id uint64) {
	id = g.s.GetNodeCount()
	*g.s.NodeCount++
	return id
}

// Nodes returns all the nodes in the Graph
func (g *Graph) Nodes() []*Node {
	n := make([]*Node, len(g.nodes))
	ctr := 0
	for _, node := range g.nodes {
		n[ctr] = node
		ctr++
	}
	return n
}

// DecodeGraph hydrates a graph from a serialized format (returned by Bytes()).
func DecodeGraph(data []byte) (*Graph, error) {
	sg, err := DecodeSerialGraph(data)
	if err != nil {
		return nil, err
	}
	g := New(false) // Don't care about directed because it's stored on s (assigned below)
	g.s = sg

	// Hydrate Graph from SerialGraph
	for _, node := range sg.Nodes {
		g.nodes[*node.Id] = node

		// initialize node adjacency map
		mm := g.edges[*node.Id]
		if mm == nil {
			mm = make(map[uint64]bool)
			g.edges[*node.Id] = mm
		}

		// populate node adjacency map
		for _, adjID := range node.GetAdjacent() {
			mm[adjID] = true
		}
	}
	return g, nil
}

// Bytes flattens a graph to a flat file format
func (g *Graph) Bytes() ([]byte, error) {
	// TODO: use smaller numbers for encoding...
	return g.s.Bytes()
}

// func main() {
// 	log.Println("Do stuff...")
//
// 	graph := New(false)
// 	a := graph.Add("http://super-stupid-long-url.com/more-crap-over-here1", NodeType_UNKNOWN, 0)
// 	b := graph.Add("http://super-stupid-long-url.com/more-crap-over-here2", NodeType_UNKNOWN, 0)
// 	graph.Connect(a, b, 0)
//
// 	// Compress
// 	bits, err := graph.Bytes()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Decompress
// 	result, err := DecodeGraph(bits)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Compare
// 	log.Printf("Message (%d): %q", len(bits), string(bits))
// 	log.Printf("Digit:\n%v\n%v", graph, result)
// 	log.Printf("Nodes:\n%v\n%v", graph.s.Nodes, result.s.Nodes)
// }
