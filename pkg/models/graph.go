package models

// GraphData is the top-level response for GET /api/graph.
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// Node represents a file in the dependency graph.
type Node struct {
	ID    string `json:"id"`              // relative file path
	Name  string `json:"name"`            // filename only
	Group int    `json:"group"`           // language group for coloring
	Val   int    `json:"val,omitempty"`   // node size (line count)
}

// Language group constants.
const (
	GroupGo     = 1
	GroupRust   = 2
	GroupJS     = 3
	GroupTS     = 4
	GroupPython = 5
	GroupJava   = 6
	GroupOther  = 99
)

// Link represents a dependency between two files.
type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}
