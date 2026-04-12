// samples/lipgloss demonstrates two integration points between Lipgloss and
// Tizzy:
//
//  1. tzlipgloss.Style() — convert a lipgloss.Style into a tz.Style so your
//     Lipgloss design tokens drive Tizzy-native components (Text, Button, …).
//
//  2. tz.NewANSIText() — embed the output of lipgloss.Render() (or any
//     ANSI-escaped string) directly inside a Tizzy layout tree.
//
// Run: go run ./samples/lipgloss
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
	"github.com/pavelgj/tizzy-go/tz"
	"github.com/pavelgj/tizzy-go/tzlipgloss"
)

// ── Design tokens ────────────────────────────────────────────────────────────
// Defined once with Lipgloss, used in both Lipgloss-rendered content (via
// ANSIText) and Tizzy-native components (via tzlipgloss.Style).

var (
	pink    = lipgloss.Color("#FF71EF")
	fuchsia = lipgloss.Color("#E637BE")
	purple  = lipgloss.Color("#9B50E7")
	cream   = lipgloss.Color("#FFFDF5")
	muted   = lipgloss.Color("#9B9B9B")
	teal    = lipgloss.Color("#00B4D8")
	green   = lipgloss.Color("#57CC99")
	amber   = lipgloss.Color("#FFBA08")
	dimBg   = lipgloss.Color("#1C1C2B")
	cardBg  = lipgloss.Color("#252538")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(pink)
	subStyle   = lipgloss.NewStyle().Faint(true).Foreground(cream)
	accentStyle = lipgloss.NewStyle().Bold(true).Foreground(fuchsia)
	linkStyle   = lipgloss.NewStyle().Foreground(teal).Underline(true)
	codeStyle   = lipgloss.NewStyle().Foreground(amber).Background(dimBg).Padding(0, 1)
	goodStyle   = lipgloss.NewStyle().Foreground(green)
	badStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Strikethrough(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		tab, setTab := tz.UseState(ctx, 0)
		confirmed, setConfirmed := tz.UseState(ctx, false)
		showModal, setShowModal := tz.UseState(ctx, false)

		tabs := []struct{ id, label string }{
			{"overview", "Overview"},
			{"adapter", "Style Adapter"},
			{"ansitext", "ANSIText"},
			{"mixed", "Mixed"},
		}

		// ── Tab bar ─────────────────────────────────────────────────────
		var tabNodes []tz.Node
		for i, t := range tabs {
			i := i
			label := " " + t.label + " "
			style := tz.Style{
				Focusable: true,
				ID:        "tab-" + t.id,
				Border:    true,
				Color:     tcell.ColorGray,
			}
			if i == tab {
				style.Color = tcell.ColorWhite
				style.Background = tcell.ColorDarkMagenta
			}
			tabNodes = append(tabNodes, tz.NewButton(style, label, func() { setTab(i) }))
		}
		tabBar := tz.NewBox(tz.Style{FlexDirection: "row", Margin: tz.Margin{Bottom: 1}}, tabNodes...)

		// ── Tab content ──────────────────────────────────────────────────
		var content tz.Node
		switch tab {

		// ────────────────────────────────────────────────────────────────
		// OVERVIEW — lipgloss-rendered logo + description side by side
		// ────────────────────────────────────────────────────────────────
		case 0:
			logo := logoBlock()
			desc := tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Left: 4}},
				// tzlipgloss.Style() drives Tizzy Text nodes
				tz.NewText(tzlipgloss.Style(titleStyle), "Style Definitions for Nice Terminal Layouts"),
				tz.NewANSIText(tz.Style{}, subStyle.Render("From Charm — ")+linkStyle.Render("https://github.com/charmbracelet/lipgloss")),
				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
					"This sample shows two integration strategies:"),
				tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, fmt.Sprintf(
					"  %s  render Lipgloss styles on tz.Text, tz.Button, …\n"+
						"  %s  embed lipgloss.Render() output via tz.NewANSIText",
					accentStyle.Render("tzlipgloss.Style()"),
					accentStyle.Render("tz.NewANSIText()"),
				)),
			)
			content = tz.NewBox(tz.Style{FlexDirection: "row"}, logo, desc)

		// ────────────────────────────────────────────────────────────────
		// STYLE ADAPTER — tzlipgloss.Style() on Tizzy-native components
		// ────────────────────────────────────────────────────────────────
		case 1:
			confirm := ""
			if confirmed {
				confirm = "  ← pressed!"
			}
			content = tz.NewBox(
				tz.Style{FlexDirection: "column"},
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					"tzlipgloss.Style() converts Lipgloss design tokens → tz.Style."),
				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Bottom: 1}},
					"Text attributes (bold, italic, underline, …) and fg/bg colors are mapped."),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Bottom: 1}},
					labeledSwatch("titleStyle", titleStyle.Render("Lip Gloss")),
					labeledSwatch("accentStyle", accentStyle.Render("Lip Gloss")),
					labeledSwatch("linkStyle", linkStyle.Render("Lip Gloss")),
					labeledSwatch("goodStyle", goodStyle.Render("Lip Gloss")),
					labeledSwatch("badStyle", badStyle.Render("Lip Gloss")),
				),

				// A Tizzy Text node styled via the adapter (not via ANSIText):
				tz.NewText(tzlipgloss.Style(titleStyle), "This tz.Text is styled via tzlipgloss.Style(titleStyle)"),
				tz.NewText(tzlipgloss.Style(subStyle), "This one uses subStyle  — faint + cream foreground"),
				tz.NewText(tzlipgloss.Style(linkStyle), "This one uses linkStyle — teal + underline"),

				// A button whose label is styled with the adapter:
				tz.NewBox(tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					tz.NewButton(
						tz.Style{
							Focusable: true, Border: true, ID: "btn-confirm",
							Color:     tzlipgloss.Color(fuchsia),
							TextAttrs: tcell.AttrBold,
						},
						" Confirm ",
						func() { setConfirmed(!confirmed) },
					),
					tz.NewText(tzlipgloss.Style(goodStyle), confirm),
				),

				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
					"Layout (padding/margin/border/width) is always set via tz.Style — the adapter skips those."),
			)

		// ────────────────────────────────────────────────────────────────
		// ANSITEXT — lipgloss.Render() output embedded as-is
		// ────────────────────────────────────────────────────────────────
		case 2:
			content = tz.NewBox(
				tz.Style{FlexDirection: "column"},
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					"tz.NewANSIText() renders the raw ANSI output of lipgloss.Render()."),
				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Bottom: 1}},
					"Tizzy parses the escape sequences and writes styled runes directly to the grid."),

				// Inline badges rendered by Lipgloss, laid out by Tizzy:
				tz.NewText(tz.Style{Color: tcell.ColorGray}, "Inline badges (each is a separate ANSIText node in a row box):"),
				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1, Bottom: 1}},
					badge("INFO", "#0077FF"),
					badge("WARN", "#FF9900"),
					badge("ERR", "#FF3333"),
					badge("OK", "#22CC66"),
					badge("NEW", "#AA44FF"),
				),

				// Code block: lipgloss renders the box; Tizzy lays it out
				tz.NewText(tz.Style{Color: tcell.ColorGray}, "Code block rendered entirely by Lipgloss:"),
				tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, codeBlock()),

				// Multi-paragraph description cards:
				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
					"Description cards — Lipgloss renders the styled text, Tizzy handles the row layout:"),
				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					descCard("Interoperability", "40",
						"Pass lipgloss.Render() output directly to tz.NewANSIText. "+
							"No conversion needed — tizzy parses the ANSI sequences."),
					tz.NewBox(tz.Style{Margin: tz.Margin{Left: 2}}, nil),
					descCard("Zero coupling", "40",
						"The core tz package has no dependency on Lipgloss. "+
							"Only import tzlipgloss when you need the style adapter."),
				),
			)

		// ────────────────────────────────────────────────────────────────
		// MIXED — both techniques side by side; modal demo
		// ────────────────────────────────────────────────────────────────
		case 3:
			// Checklist: mix of tz.Text + ANSIText
			fruits := []struct {
				name    string
				checked bool
			}{
				{"Grapefruit", true},
				{"Yuzu", true},
				{"Citron", false},
				{"Kumquat", false},
				{"Pomelo", false},
			}
			var fruitNodes []tz.Node
			for _, f := range fruits {
				icon := dimStyle.Render("◦")
				nameStyle := tz.Style{Color: tcell.ColorGray}
				if f.checked {
					icon = goodStyle.Render("✓")
					nameStyle = tz.Style{Color: tcell.ColorWhite}
				}
				fruitNodes = append(fruitNodes, tz.NewBox(
					tz.Style{FlexDirection: "row"},
					tz.NewANSIText(tz.Style{}, icon),
					tz.NewText(tz.Style{Color: nameStyle.Color, Margin: tz.Margin{Left: 1}}, f.name),
				))
			}

			// Vendors: mix of regular and ANSI-styled items
			vendors := []struct{ name, color string }{
				{"Glossier", ""},
				{"Claire's Boutique", ""},
				{"Nyx", "#FF71EF"},
				{"Mac", ""},
				{"Milk", "#57CC99"},
			}
			var vendorNodes []tz.Node
			for _, v := range vendors {
				var node tz.Node
				if v.color != "" {
					node = tz.NewANSIText(tz.Style{}, goodStyle.Render("✓")+" "+
						lipgloss.NewStyle().Foreground(lipgloss.Color(v.color)).Bold(true).Render(v.name))
				} else {
					node = tz.NewText(tz.Style{Color: tcell.ColorGray}, "  "+v.name)
				}
				vendorNodes = append(vendorNodes, node)
			}

			modalContent := tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewANSIText(tz.Style{}, accentStyle.Render("Are you sure you want to eat marmalade?")),
				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}, JustifyContent: "center"},
					tz.NewButton(tz.Style{
						Focusable: true, ID: "modal-yes", Border: true,
						Color: tcell.ColorBlack, Background: tcell.ColorDarkMagenta,
					}, " Yes ", func() { setShowModal(false) }),
					tz.NewButton(tz.Style{
						Focusable: true, ID: "modal-no", Border: true, Margin: tz.Margin{Left: 2},
					}, " Maybe ", func() { setShowModal(false) }),
				),
			)

			leftCol := tz.NewBox(
				tz.Style{FlexDirection: "column", Margin: tz.Margin{Right: 4}},
				tz.NewANSIText(tz.Style{}, accentStyle.Render("Citrus Fruits to Try")),
				tz.NewBox(tz.Style{FlexDirection: "column", Margin: tz.Margin{Top: 1}}, fruitNodes...),
				tz.NewButton(tz.Style{
					Focusable: true, ID: "btn-modal", Border: true, Margin: tz.Margin{Top: 1},
					Color: tcell.ColorWhite, Background: tcell.ColorDarkMagenta,
				}, " Open Modal ", func() { setShowModal(true) }),
			)

			rightCol := tz.NewBox(
				tz.Style{FlexDirection: "column"},
				tz.NewANSIText(tz.Style{}, accentStyle.Render("Actual Lip Gloss Vendors")),
				tz.NewBox(tz.Style{FlexDirection: "column", Margin: tz.Margin{Top: 1}}, vendorNodes...),
				tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, colorSwatchBlock()),
			)

			content = tz.NewBox(
				tz.Style{FlexDirection: "column"},
				tz.NewBox(tz.Style{FlexDirection: "row"}, leftCol, rightCol),
				tz.NewModal(ctx, tz.Style{
					Background: tcell.ColorBlack,
					Border:     true,
					Width:      48,
				}, modalContent, showModal),
			)
		}

		// ── Root frame ──────────────────────────────────────────────────
		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Left: 3, Right: 3, Bottom: 1},
				FillWidth:     true,
				FillHeight:    true,
			},
			tz.NewANSIText(tz.Style{}, titleStyle.Render("Lipgloss × Tizzy")),
			tabBar,
			tz.NewBox(
				tz.Style{Border: true, Padding: tz.Padding{Left: 2, Right: 2, Top: 1, Bottom: 1}, FillWidth: true},
				content,
			),
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
				"Tab to cycle focus • Enter/Click to interact • Ctrl+C to exit"),
		)
	}

	if err := app.Run(render, nil); err != nil {
		log.Fatal(err)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// logoBlock renders the stacked "Lip Gloss" logo as an ANSIText node.
// Each copy uses a slightly different shade to create a drop-shadow effect.
func logoBlock() tz.Node {
	shades := []string{"#C154D0", "#D45FDF", "#E46AEE", "#EE80F5", "#F599FB", "#FFB3FF"}
	var lines []string
	for _, c := range shades {
		lines = append(lines, lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(c)).
			Padding(0, 1).
			Background(lipgloss.Color("#1C0025")).
			Render("Lip Gloss"))
	}
	return tz.NewANSIText(tz.Style{}, strings.Join(lines, "\n"))
}

// badge renders a small colored label as a single ANSIText node.
func badge(label, hexColor string) tz.Node {
	rendered := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color(hexColor)).
		Padding(0, 1).
		MarginRight(1).
		Render(label)
	return tz.NewANSIText(tz.Style{}, rendered)
}

// codeBlock renders a small code sample using lipgloss's border/padding.
func codeBlock() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFBA08")).
		Background(lipgloss.Color("#1C1C2B")).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444466")).
		Render("tz.NewANSIText(style, lipgloss.Render(...))")
}

// descCard wraps a paragraph in a Lipgloss-styled border and returns an
// ANSIText node so Tizzy lays it out without re-wrapping the content.
func descCard(title, width, body string) tz.Node {
	content := lipgloss.NewStyle().
		Width(lipgloss.Width(title)+4).
		Render(body)
	_ = content

	w := 36
	rendered := lipgloss.NewStyle().
		Width(w).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444466")).
		Foreground(lipgloss.Color("#CCCCCC")).
		Render(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E46AEE")).Render(title) + "\n\n" + body)
	return tz.NewANSIText(tz.Style{}, rendered)
}

// labeledSwatch renders a Lipgloss-styled word below a small label, packaged
// as a column box so the label and swatch stay together.
func labeledSwatch(name, rendered string) tz.Node {
	return tz.NewBox(
		tz.Style{FlexDirection: "column", Margin: tz.Margin{Right: 3}},
		tz.NewANSIText(tz.Style{}, rendered),
		tz.NewText(tz.Style{Color: tcell.ColorGray}, name),
	)
}

// colorSwatchBlock builds a small 4×4 true-color gradient block using Lipgloss.
func colorSwatchBlock() string {
	var rows []string
	for row := 0; row < 4; row++ {
		var cols []string
		for col := 0; col < 8; col++ {
			r := 50 + col*25
			g := 80 + row*40
			b := 200 - col*15
			cols = append(cols, lipgloss.NewStyle().
				Background(lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))).
				Render("  "))
		}
		rows = append(rows, strings.Join(cols, ""))
	}
	return strings.Join(rows, "\n")
}
