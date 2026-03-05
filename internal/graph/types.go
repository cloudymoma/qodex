package graph

import "qodex/pkg/models"

// Builder constructs a GraphData from parsed files and dependencies.
type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

// Data holds the current in-memory graph state.
type Data struct {
	Graph *models.GraphData
	Tree  []*models.TreeNode
}

func NewData() *Data {
	return &Data{
		Graph: &models.GraphData{
			Nodes: []models.Node{},
			Links: []models.Link{},
		},
		Tree: []*models.TreeNode{},
	}
}
