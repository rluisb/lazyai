package components

import (
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/tui/theme"
)

// TreeNode represents a node in the display tree.
type TreeNode struct {
	Name     string
	Children []TreeNode
	Status   string // optional status indicator (e.g. "installed", "modified", "missing")
}

// statusToIndicator maps status strings to themed indicators.
func statusToIndicator(status string) string {
	switch status {
	case "installed", "ok", "done":
		return theme.SuccessLabel("")
	case "modified", "changed", "dirty":
		return theme.StatusModified("")
	case "missing", "error", "failed":
		return theme.StatusMissing("")
	case "conflict":
		return theme.StatusConflict("")
	case "pending":
		return theme.StatusPending("")
	default:
		return ""
	}
}

// renderNode recursively renders a single TreeNode with indentation.
func renderNode(node TreeNode, indent int, isLast bool, prefix string) string {
	var sb strings.Builder

	// Build the connector.
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Build the display line.
	name := node.Name
	if prefix == "" {
		// Root node — no connector.
		sb.WriteString(theme.Title(name))
	} else {
		sb.WriteString(prefix)
		sb.WriteString(connector)
		sb.WriteString(name)
	}

	// Append status indicator.
	if node.Status != "" && node.Status != "none" {
		indicator := statusToIndicator(node.Status)
		if indicator != "" {
			sb.WriteString(" ")
			sb.WriteString(indicator)
		}
	}

	sb.WriteString("\n")

	// Build prefix for children.
	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	} else {
		// Root level — no vertical prefix yet.
		// nth-root children use "" as base.
	}

	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		sb.WriteString(renderNode(child, indent+1, isChildLast, childPrefix))
	}

	return sb.String()
}

// RenderTree renders a TreeNode as a styled tree string.
// Pass indent=0 for the initial call.
func RenderTree(root TreeNode, indent int) string {
	return renderNode(root, indent, true, "")
}
