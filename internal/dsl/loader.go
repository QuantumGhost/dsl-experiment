package dsl

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Load parses a YAML file into a Workflow struct.
func Load(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&wf); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &wf, nil
}

// UnmarshalYAML implements custom decoding for the NodeList.
func (nl *NodeList) UnmarshalYAML(value *yaml.Node) error {
	var rawNodes []yaml.Node
	if err := value.Decode(&rawNodes); err != nil {
		return err
	}

	var nodes []Node
	for _, raw := range rawNodes {
		// 1. Decode into a wrapper to peek at the "type" field
		var wrapper struct {
			Type NodeType `yaml:"type"`
		}
		if err := raw.Decode(&wrapper); err != nil {
			return err
		}

		// 2. Instantiate the correct concrete type
		var target Node
		switch wrapper.Type {
		case NodeGenerate:
			target = &GenerateNode{}
		case NodeTool:
			target = &ToolNode{}
		case NodeParallel:
			target = &ParallelNode{}
		case NodeSelector:
			target = &SelectorNode{}
		case NodeGroup:
			target = &GroupNode{}
		case NodeNoop:
			target = &NoopNode{}
		default:
			return fmt.Errorf("unknown node type: %s", wrapper.Type)
		}

		// 3. Decode into the concrete type
		if err := raw.Decode(target); err != nil {
			return err
		}
		nodes = append(nodes, target)
	}

	*nl = nodes
	return nil
}

// UnmarshalYAML for ParallelNode to handle reducer field which is also of type Node
func (p *ParallelNode) UnmarshalYAML(value *yaml.Node) error {
	// Decode everything except reducer first
	type Alias struct {
		BaseNode `yaml:",inline"`
		Branches []Branch  `yaml:"branches,omitempty"`
		Reducer  yaml.Node `yaml:"reducer,omitempty"`
	}

	var aux Alias
	if err := value.Decode(&aux); err != nil {
		return err
	}

	p.BaseNode = aux.BaseNode
	p.Branches = aux.Branches

	// If reducer exists, decode it using our Node unmarshaler logic
	if aux.Reducer.Kind != 0 {
		// Get the type field
		var wrapper struct {
			Type NodeType `yaml:"type"`
		}
		if err := aux.Reducer.Decode(&wrapper); err != nil {
			return err
		}

		// Instantiate the correct concrete type
		var target Node
		switch wrapper.Type {
		case NodeGenerate:
			target = &GenerateNode{}
		case NodeTool:
			target = &ToolNode{}
		case NodeParallel:
			target = &ParallelNode{}
		case NodeSelector:
			target = &SelectorNode{}
		case NodeGroup:
			target = &GroupNode{}
		case NodeNoop:
			target = &NoopNode{}
		default:
			return fmt.Errorf("unknown reducer node type: %s", wrapper.Type)
		}

		if err := aux.Reducer.Decode(target); err != nil {
			return err
		}
		p.Reducer = target
	}

	return nil
}

// --- Validation Logic ---

var idPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validate(wf *Workflow) error {
	if wf.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	if len(wf.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// 1. Collect all IDs and check uniqueness
	ids := make(map[string]bool)
	if err := collectIDs(wf.Nodes, ids); err != nil {
		return err
	}

	// 2. Validate each node (Required fields, Next pointers)
	for i, node := range wf.Nodes {
		if err := validateNode(node, ids); err != nil {
			return fmt.Errorf("node[%d] (%s): %w", i, node.GetID(), err)
		}
	}

	return nil
}

func collectIDs(nodes []Node, ids map[string]bool) error {
	for _, node := range nodes {
		id := node.GetID()
		if id == "" {
			return fmt.Errorf("node has empty ID")
		}
		if !idPattern.MatchString(id) {
			return fmt.Errorf("invalid node ID format: %s", id)
		}
		if ids[id] {
			return fmt.Errorf("duplicate node ID: %s", id)
		}
		ids[id] = true

		// Recursive collection for nested nodes
		switch n := node.(type) {
		case *ParallelNode:
			for _, branch := range n.Branches {
				if err := collectIDs(branch.Nodes, ids); err != nil {
					return err
				}
			}
			if n.Reducer != nil {
				if err := collectIDs([]Node{n.Reducer}, ids); err != nil {
					return err
				}
			}
		case *GroupNode:
			if err := collectIDs(n.Nodes, ids); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateNode(node Node, allIDs map[string]bool) error {
	// Check Next pointer
	if next := node.GetNext(); next != "" {
		if !allIDs[next] {
			return fmt.Errorf("next pointer '%s' not found", next)
		}
	}

	// Type-specific validation
	switch n := node.(type) {
	case *GenerateNode:
		if n.Model == "" {
			return fmt.Errorf("missing required field: model")
		}
	case *SelectorNode:
		if len(n.Options) == 0 {
			return fmt.Errorf("selector must have at least one option")
		}
		for _, opt := range n.Options {
			if opt.Next == "" {
				return fmt.Errorf("option case '%s' missing next pointer", opt.Case)
			}
			if !allIDs[opt.Next] {
				return fmt.Errorf("option next pointer '%s' not found", opt.Next)
			}
		}
	case *ParallelNode:
		if len(n.Branches) == 0 {
			return fmt.Errorf("parallel node must have at least one branch")
		}
		// Validate nested nodes
		for _, branch := range n.Branches {
			for _, subNode := range branch.Nodes {
				if err := validateNode(subNode, allIDs); err != nil {
					return err
				}
			}
		}
		if n.Reducer != nil {
			if err := validateNode(n.Reducer, allIDs); err != nil {
				return err
			}
		}
	case *GroupNode:
		if len(n.Nodes) == 0 {
			return fmt.Errorf("group node must have at least one node")
		}
		for _, subNode := range n.Nodes {
			if err := validateNode(subNode, allIDs); err != nil {
				return err
			}
		}
	}
	return nil
}
