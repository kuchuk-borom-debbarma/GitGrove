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

// FindNode finds a node by repository name in the hierarchy.
// Returns nil if not found.
func (li *LinkInfo) FindNode(repoName string) *TreeNode {
	// Build a flat map for quick lookup
	var findInNodes func([]*TreeNode) *TreeNode
	findInNodes = func(nodes []*TreeNode) *TreeNode {
		for _, node := range nodes {
			if node.State.Repo.Name == repoName {
				return node
			}
			if found := findInNodes(node.Children); found != nil {
				return found
			}
		}
		return nil
	}
	return findInNodes(li.Roots)
}

// String returns a beautiful tree representation of the hierarchy
func (ld *LinkInfo) String(currentRepo string) string {
	var sb strings.Builder
	for _, root := range ld.Roots {
		ld.printNode(&sb, root, "", true, currentRepo)
	}
	return sb.String()
}

func (ld *LinkInfo) printNode(sb *strings.Builder, node *TreeNode, prefix string, isLast bool, currentRepo string) {
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└── ")
		prefix += "    "
	} else {
		sb.WriteString("├── ")
		prefix += "│   "
	}

	name := node.State.Repo.Name
	if name == currentRepo {
		sb.WriteString(fmt.Sprintf("* %s (%s) [CURRENT]", name, node.State.Repo.Path))
	} else {
		sb.WriteString(fmt.Sprintf("%s (%s)", name, node.State.Repo.Path))
	}

	if !node.State.PathExists && name != currentRepo {
		// Only show MISSING if it's not the current repo (which by definition exists here)
		// Actually, if we are in a flattened view, other repos ARE missing.
		// But we want to show the hierarchy regardless.
		sb.WriteString(" [MISSING]")
	}

	sb.WriteString("\n")

	for i, child := range node.Children {
		ld.printNode(sb, child, prefix, i == len(node.Children)-1, currentRepo)
	}
}
