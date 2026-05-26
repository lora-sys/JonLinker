package engine

import (
	"context"
	"fmt"
	"log"

	"joblinker/internal/engine/nodes"
)

// NodeType identifies the type of processing node in the graph.
type NodeType string

const (
	NodeTypeAI       NodeType = "ai"
	NodeTypeTool     NodeType = "tool"
	NodeTypeMemory   NodeType = "memory"
	NodeTypeConfirm  NodeType = "confirm"
)

// Edge connects two nodes in the graph.
type Edge struct {
	From NodeType
	To   NodeType
	// Condition is an optional routing function.
	// If nil, the edge is always taken.
	Condition func(output map[string]interface{}) bool
}

// Graph is a directed graph of processing nodes with conditional edges.
type Graph struct {
	nodes map[NodeType]Processor
	edges []Edge
	entry NodeType
}

// Processor is implemented by all node types in the graph.
type Processor interface {
	Process(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// NewGraph creates an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[NodeType]Processor),
		edges: []Edge{},
	}
}

// AddNode registers a processing node.
func (g *Graph) AddNode(nt NodeType, proc Processor) {
	g.nodes[nt] = proc
}

// AddEdge adds a directed edge between nodes.
func (g *Graph) AddEdge(from, to NodeType, condition func(map[string]interface{}) bool) {
	g.edges = append(g.edges, Edge{
		From:      from,
		To:        to,
		Condition: condition,
	})
}

// SetEntry sets the entry point of the graph.
func (g *Graph) SetEntry(nt NodeType) {
	g.entry = nt
}

// Execute runs the graph from the entry node, following edges based on conditions.
// Returns the output of the final node reached.
func (g *Graph) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	if g.entry == "" {
		return nil, fmt.Errorf("graph has no entry point")
	}

	current := g.entry
	var output map[string]interface{}
	var err error

	maxSteps := 20
	for step := 0; step < maxSteps; step++ {
		proc, ok := g.nodes[current]
		if !ok {
			return nil, fmt.Errorf("node %s not registered", current)
		}

		log.Printf("[Graph] executing node: %s (step %d)", current, step)
		output, err = proc.Process(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("node %s failed: %w", current, err)
		}

		// Merge output back into input for next node
		for k, v := range output {
			input[k] = v
		}

		// Find next node via edges
		next := g.findNext(current, output)
		if next == "" {
			// Terminal node - return output
			return output, nil
		}
		current = next
	}

	return output, fmt.Errorf("graph exceeded max steps (%d)", maxSteps)
}

func (g *Graph) findNext(current NodeType, output map[string]interface{}) NodeType {
	for _, edge := range g.edges {
		if edge.From != current {
			continue
		}
		if edge.Condition == nil || edge.Condition(output) {
			return edge.To
		}
	}
	return ""
}

// NodeRegistry creates a standard agent conversation graph.
type NodeRegistry struct {
	AINode     *nodes.AINode
	ToolNode   *nodes.ToolNode
	MemoryNode *nodes.MemoryNode
	ConfirmNode *nodes.ConfirmNode
}

// BuildStandardGraph builds the standard agent conversation graph:
// Memory → AI → Tool (if tool intent) or Confirm (if confirm action) → AI (final)
func (nr *NodeRegistry) BuildStandardGraph() *Graph {
	g := NewGraph()

	g.AddNode(NodeTypeMemory, nr.MemoryNode)
	g.AddNode(NodeTypeAI, nr.AINode)
	g.AddNode(NodeTypeTool, nr.ToolNode)
	g.AddNode(NodeTypeConfirm, nr.ConfirmNode)

	// Memory → AI (always)
	g.AddEdge(NodeTypeMemory, NodeTypeAI, nil)

	// AI → Tool (if tool call detected)
	g.AddEdge(NodeTypeAI, NodeTypeTool, func(output map[string]interface{}) bool {
		intent, _ := output["intent"].(string)
		return intent == "TOOL_CALL"
	})

	// AI → Confirm (if confirm/accept/reject action)
	g.AddEdge(NodeTypeAI, NodeTypeConfirm, func(output map[string]interface{}) bool {
		intent, _ := output["intent"].(string)
		return intent == "SCHEDULE_INTERVIEW" || intent == "OFFER_ACCEPTED" || intent == "OFFER_DECLINED"
	})

	// Tool → AI (back for follow-up)
	g.AddEdge(NodeTypeTool, NodeTypeAI, nil)

	// Confirm → AI (back for follow-up)
	g.AddEdge(NodeTypeConfirm, NodeTypeAI, nil)

	g.SetEntry(NodeTypeMemory)

	return g
}
