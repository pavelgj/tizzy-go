package splotch

import "strings"

// LayoutResult holds the calculated position and size of a node.
type LayoutResult struct {
	Node     Node
	X, Y     int
	W, H     int
	Children []LayoutResult
}

// Constraints defines the maximum size allowed for a node.
type Constraints struct {
	MaxW, MaxH int
}

// Layout calculates the layout of a node and its children.
func Layout(node Node, x, y int, c Constraints) LayoutResult {
	switch n := node.(type) {
	case *Text:
		pad := n.Style.Padding
		margin := n.Style.Margin

		// Position of the border box
		boxX := x + margin.Left
		boxY := y + margin.Top

		return LayoutResult{
			Node: node,
			X:    boxX,
			Y:    boxY,
			W:    len(n.Content) + pad.Left + pad.Right,
			H:    1 + pad.Top + pad.Bottom,
		}
	case *TextInput:
		pad := n.Style.Padding
		margin := n.Style.Margin

		boxX := x + margin.Left
		boxY := y + margin.Top

		w := 0
		h := 1
		if n.Style.Multiline {
			lines := strings.Split(n.Value, "\n")
			for _, line := range lines {
				if len(line) > w {
					w = len(line)
				}
			}
			h = len(lines)
		} else {
			w = len(n.Value)
		}
		
		if n.Style.Width > 0 {
			w = n.Style.Width
		}

		borderSize := 0
		if n.Style.Border {
			borderSize = 2
		}

		return LayoutResult{
			Node: node,
			X:    boxX,
			Y:    boxY,
			W:    w + pad.Left + pad.Right + borderSize,
			H:    h + pad.Top + pad.Bottom + borderSize,
		}
	case *Box:
		borderSize := 0
		if n.Style.Border {
			borderSize = 1
		}

		pad := n.Style.Padding
		margin := n.Style.Margin

		boxX := x + margin.Left
		boxY := y + margin.Top

		res := LayoutResult{
			Node: node,
			X:    boxX,
			Y:    boxY,
			W:    0,
			H:    0,
		}

		// Children start after border and padding
		curX := boxX + borderSize + pad.Left
		curY := boxY + borderSize + pad.Top

		contentW := 0
		contentH := 0

		childConstraints := Constraints{
			MaxW: c.MaxW - (borderSize * 2) - pad.Left - pad.Right,
			MaxH: c.MaxH - (borderSize * 2) - pad.Top - pad.Bottom,
		}
		if childConstraints.MaxW < 0 {
			childConstraints.MaxW = 0
		}
		if childConstraints.MaxH < 0 {
			childConstraints.MaxH = 0
		}

		for _, child := range n.Children {
			// Get child's style to read margin
			var childMargin Margin
			switch c := child.(type) {
			case *Text:
				childMargin = c.Style.Margin
			case *Box:
				childMargin = c.Style.Margin
			}

			// cRes.X will be curX + childMargin.Left
			cRes := Layout(child, curX, curY, childConstraints)
			res.Children = append(res.Children, cRes)

			if n.Style.FlexDirection == "row" {
				curX += cRes.W + childMargin.Right
				contentW += cRes.W + childMargin.Left + childMargin.Right
				if cRes.H+childMargin.Top+childMargin.Bottom > contentH {
					contentH = cRes.H + childMargin.Top + childMargin.Bottom
				}
			} else { // default to column
				curY += cRes.H + childMargin.Bottom
				contentH += cRes.H + childMargin.Top + childMargin.Bottom
				if cRes.W+childMargin.Left+childMargin.Right > contentW {
					contentW = cRes.W + childMargin.Left + childMargin.Right
				}
			}
		}

		// Total size of this box (excluding its own margin)
		res.W = contentW + (borderSize * 2) + pad.Left + pad.Right
		res.H = contentH + (borderSize * 2) + pad.Top + pad.Bottom

		// If we are centering, we take up all available space!
		if n.Style.JustifyContent == "center" {
			if n.Style.FlexDirection == "row" {
				if c.MaxW > res.W {
					res.W = c.MaxW
				}
			} else {
				if c.MaxH > res.H {
					res.H = c.MaxH
				}
			}
		}

		// Flexbox alignment
		if n.Style.JustifyContent == "center" {
			if n.Style.FlexDirection == "row" {
				remainingW := res.W - (borderSize * 2) - pad.Left - pad.Right - contentW
				if remainingW > 0 {
					shift := remainingW / 2
					for i := range res.Children {
						res.Children[i].X += shift
					}
				}
			} else { // column
				remainingH := res.H - (borderSize * 2) - pad.Top - pad.Bottom - contentH
				if remainingH > 0 {
					shift := remainingH / 2
					for i := range res.Children {
						res.Children[i].Y += shift
					}
				}
			}
		}

		return res
	}
	return LayoutResult{Node: node, X: x, Y: y, W: 0, H: 0}
}
