package tz

import "fmt"

// walkTree performs a depth-first walk of the node tree, calling fn on each
// node. Walking stops early if fn returns a non-nil error.
func walkTree(node Node, fn func(Node) error) error {
	if node == nil {
		return nil
	}
	if err := fn(node); err != nil {
		return err
	}
	if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
			if err := walkTree(child, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

// findNodeByID returns the first node in the tree whose style ID matches id.
func findNodeByID(node Node, id string) Node {
	if node == nil {
		return nil
	}
	if node.GetStyle().ID == id {
		return node
	}
	if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
			if found := findNodeByID(child, id); found != nil {
				return found
			}
		}
	}
	return nil
}

// validateUniqueIDs returns an error if any two nodes in the tree share the
// same non-empty style ID.
func validateUniqueIDs(node Node) error {
	seen := make(map[string]bool)
	return walkTree(node, func(n Node) error {
		id := n.GetStyle().ID
		if id != "" {
			if seen[id] {
				return fmt.Errorf("duplicate component ID: %s", id)
			}
			seen[id] = true
		}
		return nil
	})
}

// findFocusableIDs returns the ordered list of focusable component IDs
// reachable from node. Portal children are skipped (TrapFocus portals are
// handled separately in RenderFrame). Components implementing FocusScope
// control which of their children are included.
func findFocusableIDs(node Node, componentStates map[string]any) []string {
	var ids []string
	if node == nil {
		return ids
	}
	if f, ok := node.(Focusable); ok && f.IsFocusable() && node.GetStyle().ID != "" {
		ids = append(ids, node.GetStyle().ID)
	}
	// Portal children are not traversed for normal focus discovery.
	if _, ok := node.(*Portal); ok {
		return ids
	}
	if scope, ok := node.(FocusScope); ok {
		for _, child := range scope.FocusableChildren(componentStates) {
			ids = append(ids, findFocusableIDs(child, componentStates)...)
		}
	} else if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
			ids = append(ids, findFocusableIDs(child, componentStates)...)
		}
	}
	return ids
}

// nextFocus returns the ID that follows current in the ids slice (wrapping).
func nextFocus(current string, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	if current == "" {
		return ids[0]
	}
	for i, id := range ids {
		if id == current {
			return ids[(i+1)%len(ids)]
		}
	}
	return ids[0]
}

// prevFocus returns the ID that precedes current in the ids slice (wrapping).
func prevFocus(current string, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	if current == "" {
		return ids[len(ids)-1]
	}
	for i, id := range ids {
		if id == current {
			return ids[(i-1+len(ids))%len(ids)]
		}
	}
	return ids[len(ids)-1]
}

// findLayoutResultByID returns the LayoutResult whose node has the given ID,
// searching the result tree depth-first.
func findLayoutResultByID(res LayoutResult, id string) *LayoutResult {
	if res.Node.GetStyle().ID == id {
		return &res
	}
	for _, child := range res.Children {
		if found := findLayoutResultByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// findNodePathAt returns the path of nodes from root to the leaf node that
// contains screen coordinate (x, y), or nil if no node covers that point.
// Components implementing CustomHitTester can override child traversal.
func findNodePathAt(res LayoutResult, x, y int, componentStates map[string]any) []Node {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		if hitTester, ok := res.Node.(CustomHitTester); ok {
			return hitTester.FindNodePathAt(x, y, res, componentStates)
		}
		for _, child := range res.Children {
			if path := findNodePathAt(child, x, y, componentStates); path != nil {
				return append([]Node{res.Node}, path...)
			}
		}
		return []Node{res.Node}
	}
	return nil
}
