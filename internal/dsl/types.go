package dsl

type NodeType string
type Visibility string
type MemoryPolicy string
type ContextMode string

const (
	NodeGenerate NodeType = "generate"
	NodeTool     NodeType = "tool"
	NodeParallel NodeType = "parallel"
	NodeSelector NodeType = "selector"
	NodeGroup    NodeType = "group"
	NodeNoop     NodeType = "noop"

	VisibilityPublic    Visibility = "public"
	VisibilitySilent    Visibility = "silent"
	VisibilityEphemeral Visibility = "ephemeral"

	MemoryResultOnly MemoryPolicy = "result_only"
	MemoryFull       MemoryPolicy = "full"
	MemorySummary    MemoryPolicy = "summary"

	ContextFull   ContextMode = "full"
	ContextNone   ContextMode = "none"
	ContextWindow ContextMode = "window"
)

type NodeList []Node

type Workflow struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Nodes       NodeList `yaml:"nodes" json:"nodes"`
}

// Node is a sealed interface for all node types.
type Node interface {
	isNode()
	GetID() string
	GetType() NodeType
	GetNext() string
}

// BaseNode contains common fields.
type BaseNode struct {
	ID           string       `yaml:"id" json:"id"`
	Type         NodeType     `yaml:"type" json:"type"`
	Next         string       `yaml:"next,omitempty" json:"next,omitempty"`
	Visibility   Visibility   `yaml:"visibility,omitempty" json:"visibility,omitempty"`
	MemoryPolicy MemoryPolicy `yaml:"memory_policy,omitempty" json:"memory_policy,omitempty"`
}

func (n BaseNode) isNode()           {}
func (n BaseNode) GetID() string     { return n.ID }
func (n BaseNode) GetType() NodeType { return n.Type }
func (n BaseNode) GetNext() string   { return n.Next }

type GenerateNode struct {
	BaseNode     `yaml:",inline"`
	Model        string      `yaml:"model,omitempty" json:"model,omitempty"`
	SystemPrompt string      `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Temperature  float64     `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	ContextMode  ContextMode `yaml:"context_mode,omitempty" json:"context_mode,omitempty"`
}

type ToolNode struct {
	BaseNode   `yaml:",inline"`
	ToolName   string                 `yaml:"tool_name" json:"tool_name"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

type ParallelNode struct {
	BaseNode `yaml:",inline"`
	Branches []Branch `yaml:"branches,omitempty" json:"branches,omitempty"`
	Reducer  Node     `yaml:"reducer,omitempty" json:"reducer,omitempty"`
}

type SelectorNode struct {
	BaseNode     `yaml:",inline"`
	Model        string   `yaml:"model,omitempty" json:"model,omitempty"`
	SystemPrompt string   `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	Options      []Option `yaml:"options,omitempty" json:"options,omitempty"`
}

type GroupNode struct {
	BaseNode `yaml:",inline"`
	Nodes    NodeList `yaml:"nodes,omitempty" json:"nodes,omitempty"`
}

type NoopNode struct {
	BaseNode `yaml:",inline"`
}

type Branch struct {
	Nodes NodeList `yaml:"nodes" json:"nodes"`
}

type Option struct {
	Case string `yaml:"case" json:"case"`
	Next string `yaml:"next" json:"next"`
}
