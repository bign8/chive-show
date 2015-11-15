package main

import (
	"errors"
	"log"

	"github.com/golang/protobuf/proto"
)

// TODO: add some graph processing functions

// Graph is the serializable graph we have all been looking for
type Graph struct {
	s     *SerialGraph
	nodes map[uint64]*Node // Optimal lookup with pointers goes here
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
	}
}

// Add creates and adds a node to the graph
func (g *Graph) Add(value string, ttype NodeType, weight int64) *Node {
	n := &Node{
		Id:       proto.Uint64(g.genNodeID()),
		Value:    proto.String(value),
		Weight:   proto.Int64(weight),
		Type:     ttype.Enum(),
		Adjacent: make([]uint64, 0),
	}
	g.nodes[*n.Id] = n
	g.s.Nodes = append(g.s.Nodes, n)
	return n
}

// Connect connects nodes to and from with an edge of weight w
func (g *Graph) Connect(to, from *Node, weight int64) error {
	if to == nil || from == nil {
		return errors.New("Cannot add edge to nil node")
	}
	from.Adjacent = append(from.Adjacent, *to.Id) // Directed edge
	from.Weights = append(from.Weights, weight)

	if !g.s.GetDirected() { // UnDirected edge (return trip)
		to.Adjacent = append(to.Adjacent, *from.Id)
		to.Weights = append(to.Weights, weight)
	}
	return nil
}

func (g *Graph) genNodeID() (id uint64) {
	id = g.s.GetNodeCount()
	*g.s.NodeCount++
	return id
}

// DecodeGraph hydrates a graph from a serialized format (returned by Bytes()).
func DecodeGraph(data []byte) (*Graph, error) {
	sg, err := DecodeSerialGraph(data)
	if err != nil {
		return nil, err
	}
	g := &Graph{sg, make(map[uint64]*Node)}

	// Hydrate Graph from SerialGraph
	for _, node := range sg.Nodes {
		g.nodes[*node.Id] = node
	}
	return g, nil
}

// Bytes flattens a graph to a flat file format
func (g *Graph) Bytes() ([]byte, error) {
	// TODO: use smaller numbers for encoding...
	return g.s.Bytes()
}

func main() {
	log.Println("Do stuff...")

	graph := New(false)
	a := graph.Add("http://super-stupid-long-url.com/more-crap-over-here1", NodeType_UNKNOWN, 0)
	b := graph.Add("http://super-stupid-long-url.com/more-crap-over-here2", NodeType_UNKNOWN, 0)
	graph.Connect(a, b, 0)

	// Compress
	bits, err := graph.Bytes()
	if err != nil {
		panic(err)
	}

	// Decompress
	result, err := DecodeGraph(bits)
	if err != nil {
		panic(err)
	}

	// Compare
	log.Printf("Message (%d): %q", len(bits), string(bits))
	log.Printf("Digit:\n%v\n%v", graph, result)
	log.Printf("Nodes:\n%v\n%v", graph.s.Nodes, result.s.Nodes)
}
