package graph

import "errors"

// TODO: add some graph processing functions

// NodeID is a graph identifier
type NodeID uint64

// Graph is the serializable graph we have all been looking for
type Graph struct {
	s     *SerialGraph
	dupes map[NodeType]map[string]NodeID // type > value > node
}

// New creates a new Graph
func New(isDirected bool) *Graph {
	return &Graph{
		s: &SerialGraph{
			Nodes:     make(map[uint64]*Node),
			Directed:  isDirected,
			NodeCount: 0,
		},
		dupes: make(map[NodeType]map[string]NodeID),
	}
}

// Get returns an associated node for a given ID
func (g *Graph) Get(id NodeID) *Node {
	return g.s.Nodes[uint64(id)]
}

// Add creates and adds a node to the graph
func (g *Graph) Add(value string, ttype NodeType, weight int64) NodeID {

	// Check duplicate node (add weight)
	dupe := g.dupes[ttype][value]
	if dupe != 0 {
		g.Get(dupe).Weight += weight
		return dupe
	}

	// Create new node
	id := g.genNodeID()
	n := &Node{
		Value:    value,
		Weight:   weight,
		Type:     ttype,
		Adjacent: make(map[uint64]int64, 0),
	}
	g.s.Nodes[id] = n

	// Add dupe check to list
	dub, ok := g.dupes[ttype]
	if !ok {
		dub = make(map[string]NodeID)
		g.dupes[ttype] = dub
	}
	nid := NodeID(id)
	dub[value] = nid
	return nid
}

// Connect connects nodes to and from with an edge of weight w
func (g *Graph) Connect(from, to NodeID, weight int64) error {
	if to == 0 || from == 0 {
		return errors.New("Cannot add edge to nil node")
	}
	g.Get(from).Adjacent[uint64(to)] += weight // Directed edge
	if !g.s.Directed {
		g.Get(to).Adjacent[uint64(from)] += weight // UnDirected edge (return trip)
	}
	return nil
}

func (g *Graph) genNodeID() (id uint64) {
	g.s.NodeCount++
	id = g.s.NodeCount
	return id
}

// Nodes returns all the nodes in the Graph
func (g *Graph) Nodes() []*Node {
	n := make([]*Node, len(g.s.Nodes))
	ctr := 0
	for _, node := range g.s.Nodes {
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
	for id, node := range sg.Nodes {
		nn := g.dupes[node.Type]
		if nn == nil {
			nn = make(map[string]NodeID)
			g.dupes[node.Type] = nn
		}
		nn[node.Value] = NodeID(id)
	}
	return g, nil
}

// Bytes flattens a graph to a flat file format
func (g *Graph) Bytes() ([]byte, error) {
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
