package info

import (
	"fmt"
	"sort"
	"strings"
)

// LinkInfo represents the hierarchical structure of repositories.
type LinkInfo struct {
	Roots []*TreeNode
}

// TreeNode represents a single repository in the hierarchy.
type TreeNode struct {
	State    RepoState
	Children []*TreeNode
}

// GetLinkInfo builds the hierarchical structure of repositories from a RepoInfo.
func GetLinkInfo(repoInfo *RepoInfo) *LinkInfo {
	// Build map of name -> node
	nodes := make(map[string]*TreeNode)
	for name, state := range repoInfo.Repos {
		nodes[name] = &TreeNode{
			State:    state,
			Children: []*TreeNode{},
		}
	}

	// Link children to parents
	var roots []*TreeNode
	for _, node := range nodes {
		if node.State.Repo.Parent == "" {
			roots = append(roots, node)
		} else {
			if parent, ok := nodes[node.State.Repo.Parent]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// Parent missing from map (should not happen if consistent)
				// Treat as root for display purposes, maybe mark as broken link?
				// For now, just append to roots so it's visible.
				roots = append(roots, node)
			}
		}
	}

	// Sort roots and children for consistent output
	sortNodes(roots)
	for _, node := range nodes {
		sortNodes(node.Children)
	}

	return &LinkInfo{Roots: roots}
}

func sortNodes(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].State.Repo.Name < nodes[j].State.Repo.Name
	})
}

// String returns a beautiful tree representation of the hierarchy
func (ld *LinkInfo) String() string {
	var sb strings.Builder
	for _, root := range ld.Roots {
		ld.printNode(&sb, root, "", true)
	}
	return sb.String()
}

func (ld *LinkInfo) printNode(sb *strings.Builder, node *TreeNode, prefix string, isLast bool) {
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
		prefix += "    "
	} else {
		sb.WriteString("├── ")
		prefix += "│   "
	}

	sb.WriteString(fmt.Sprintf("%s (%s)", node.State.Repo.Name, node.State.Repo.Path))

	if !node.State.PathExists {
		sb.WriteString(" [MISSING]")
	}

	// Check for broken parent link (if parent is set but not found in tree logic)
	// The tree construction logic puts orphans in roots.
	// We can check if it has a parent defined but is at root level (and not empty parent)
	// However, the recursion doesn't know if it's a root because of logic or definition.
	// A simpler check: if we are printing a node, we can check its parent field.
	// But we don't have easy access to the parent node here to check if it exists.
	// The GetLinkStatus logic handles the structure.
	// If a node claims a parent "foo" but "foo" doesn't exist, it ends up as a root.
	// We can detect that case if we want, but "MISSING" path is the main request.

	sb.WriteString("\n")

	for i, child := range node.Children {
		ld.printNode(sb, child, prefix, i == len(node.Children)-1)
	}
}
