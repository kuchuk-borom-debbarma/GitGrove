package status

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kuchuk-borom-debbarma/GitGrove/core/internal/grove/model"
)

type LinkStatus struct {
	Roots []*TreeNode
}

type TreeNode struct {
	Repo     model.Repo
	Children []*TreeNode
}

func GetLinkStatus(repoStatus *RepoStatus) *LinkStatus {
	// Build map of name -> node
	nodes := make(map[string]*TreeNode)
	for name, repo := range repoStatus.Repos {
		nodes[name] = &TreeNode{
			Repo:     repo,
			Children: []*TreeNode{},
		}
	}

	// Link children to parents
	var roots []*TreeNode
	for _, node := range nodes {
		if node.Repo.Parent == "" {
			roots = append(roots, node)
		} else {
			if parent, ok := nodes[node.Repo.Parent]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Orphaned node (should not happen in valid state)
				roots = append(roots, node)
			}
		}
	}

	// Sort roots and children for consistent output
	sortNodes(roots)
	for _, node := range nodes {
		sortNodes(node.Children)
	}

	return &LinkStatus{Roots: roots}
}

func sortNodes(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Repo.Name < nodes[j].Repo.Name
	})
}

// String returns a beautiful tree representation of the hierarchy
func (ls *LinkStatus) String() string {
	var sb strings.Builder
	for _, root := range ls.Roots {
		ls.printNode(&sb, root, "", true)
	}
	return sb.String()
}

func (ls *LinkStatus) printNode(sb *strings.Builder, node *TreeNode, prefix string, isLast bool) {
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
		prefix += "    "
	} else {
		sb.WriteString("├── ")
		prefix += "│   "
	}
	sb.WriteString(fmt.Sprintf("%s (%s)\n", node.Repo.Name, node.Repo.Path))

	for i, child := range node.Children {
		ls.printNode(sb, child, prefix, i == len(node.Children)-1)
	}
}
