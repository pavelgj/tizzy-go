package tz

import (
	"flag"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func verifyVisual(t *testing.T, grid *Grid, name string) {
	t.Helper()
	goldenDir := "testdata/golden"
	failedDir := "testdata/failed"
	goldenPath := filepath.Join(goldenDir, name+".png")

	// Ensure directories exist
	os.MkdirAll(goldenDir, 0755)
	os.MkdirAll(failedDir, 0755)

	if *updateGolden {
		err := grid.DumpToPNG(goldenPath, 10, 20)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Load golden image
	goldenFile, err := os.Open(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file %s does not exist. Run with -update to create it.", goldenPath)
		}
		t.Fatal(err)
	}
	defer goldenFile.Close()
	goldenImg, err := png.Decode(goldenFile)
	if err != nil {
		t.Fatal(err)
	}

	// Generate current image in memory (or temp file)
	tempPath := filepath.Join(failedDir, "temp_"+name+".png")
	err = grid.DumpToPNG(tempPath, 10, 20)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempPath)

	tempFile, err := os.Open(tempPath)
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	currentImg, err := png.Decode(tempFile)
	if err != nil {
		t.Fatal(err)
	}

	if !compareImages(goldenImg, currentImg) {
		failPath := filepath.Join(failedDir, name+"_failed.png")
		// Copy temp to fail path
		grid.DumpToPNG(failPath, 10, 20)
		t.Fatalf("Visual regression detected for %s. New image saved to %s", name, failPath)
	}
}

func compareImages(img1, img2 image.Image) bool {
	b1 := img1.Bounds()
	b2 := img2.Bounds()
	if b1 != b2 {
		return false
	}
	for y := b1.Min.Y; y < b1.Max.Y; y++ {
		for x := b1.Min.X; x < b1.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				return false
			}
		}
	}
	return true
}

func TestGenerateListVisual(t *testing.T) {
	ctx := makeTestContext()
	items := []any{"Item 1", "Item 2", "Item 3"}

	list := NewList(
		ctx,
		Style{ID: "mylist", Width: 20, Height: 10, Border: true, Title: "List"},
		"key",
		items,
		-1,
		func(item any, index int, selected bool, cursor bool) Node {
			return NewListItem(item.(string), selected, cursor)
		},
		nil,
	)

	layout := Layout(list, 0, 0, Constraints{MaxW: 20, MaxH: 10})
	grid := NewGrid(20, 10)

	state := &ListState{SelectedIndex: 1, CursorIndex: 2}
	compStates := map[string]any{"mylist": state}

	list.Render(grid, layout, "", compStates)

	verifyVisual(t, grid, "list_layout")
}

func TestInspectStyle(t *testing.T) {
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	v := reflect.ValueOf(style)
	t.Logf("Type: %T", style)
	t.Logf("Kind: %s", v.Kind())
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	t.Logf("Concrete Type: %s", v.Type())
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			t.Logf("Field %d: %s = %v", i, v.Type().Field(i).Name, v.Field(i))
		}
	}
}

type GroceryItem struct {
	Name        string
	Description string
}

func TestGenerateGroceryListVisual(t *testing.T) {
	ctx := makeTestContext()
	items := []any{
		GroceryItem{Name: "Pocky", Description: "Expensive"},
		GroceryItem{Name: "Ginger", Description: "Exquisite"},
		GroceryItem{Name: "Plantains", Description: "Questionable"},
		GroceryItem{Name: "Honey Dew", Description: "Delectable"},
		GroceryItem{Name: "Pineapple", Description: "Kind of spicy"},
		GroceryItem{Name: "Snow Peas", Description: "Bold flavor"},
		GroceryItem{Name: "Party Gherkin", Description: "My favorite"},
	}

	list := NewList(
		ctx,
		Style{
			Width:     40,
			Height:    16,
			Border:    true,
			Title:     "Groceries",
			Focusable: true,
			ID:        "grocery-list",
		},
		"grocery-list",
		items,
		-1,
		func(item any, index int, selected bool, cursor bool) Node {
			g := item.(GroceryItem)

			bg := tcell.ColorBlack
			if cursor {
				bg = tcell.ColorDarkCyan
			}
			if selected {
				bg = tcell.ColorBlue
			}

			titleColor := tcell.ColorWhite
			descColor := tcell.ColorGray

			if cursor || selected {
				titleColor = tcell.ColorWhite
				descColor = tcell.ColorLightGray
			}

			return NewBox(
				Style{
					FlexDirection: "column",
					Background:    bg,
					FillWidth:     true,
				},
				NewText(Style{Color: titleColor, Background: bg}, g.Name),
				NewText(Style{Color: descColor, Background: bg}, g.Description),
			)
		},
		nil,
	)
	list.ItemHeight = 2

	root := NewBox(
		Style{
			Width:          50,
			Height:         20,
			JustifyContent: "center",
			FlexDirection:  "column",
		},
		NewText(Style{Color: tcell.ColorGreen}, "Use Arrows to navigate, Enter to select"),
		list,
		NewText(Style{Color: tcell.ColorGreen}, "Selected Index: -1"),
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 50, MaxH: 20})
	grid := NewGrid(50, 20)

	state := &ListState{SelectedIndex: -1, CursorIndex: 1} // Highlight Ginger
	compStates := map[string]any{"grocery-list": state}

	Render(grid, layout, "grocery-list", compStates)

	verifyVisual(t, grid, "grocery_list_layout")
}

func TestGenerateDropdownVisual(t *testing.T) {
	opts := []string{"Option 1", "Option 2", "Option 3"}

	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(40, 15)

	app := NewAppWithScreen(screen)
	app.componentStates["dropdown"] = &DropdownState{Open: true, FocusedIndex: 1}

	renderFn := func(ctx *RenderContext) Node {
		dropdown := NewDropdown(ctx, Style{ID: "dropdown", Width: 20, Color: tcell.ColorWhite, Background: tcell.ColorBlack, Border: true}, opts, 0, nil)
		return NewBox(
			Style{Width: 40, Height: 15, JustifyContent: "center"},
			dropdown,
		)
	}

	grid, _, _, _, err := app.RenderFrame(renderFn)
	if err != nil {
		t.Fatal(err)
	}

	verifyVisual(t, grid, "dropdown_layout")
}

func TestGenerateDropdownOpenAboveVisual(t *testing.T) {
	opts := []string{"Option 1", "Option 2", "Option 3"}

	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(40, 15)

	app := NewAppWithScreen(screen)
	app.componentStates["dropdown"] = &DropdownState{Open: true, FocusedIndex: 1}

	renderFn := func(ctx *RenderContext) Node {
		dropdown := NewDropdown(ctx, Style{ID: "dropdown", Width: 20, Color: tcell.ColorWhite, Background: tcell.ColorBlack, Border: true}, opts, 0, nil)
		// Place dropdown at bottom by putting a spacer box above it
		return NewBox(
			Style{Width: 40, Height: 15, FlexDirection: "column"},
			NewBox(Style{Height: 10}), // Spacer
			dropdown,
		)
	}

	grid, _, _, _, err := app.RenderFrame(renderFn)
	if err != nil {
		t.Fatal(err)
	}

	verifyVisual(t, grid, "dropdown_above_layout")
}

func TestGenerateGalleryVisual(t *testing.T) {
	ctx := makeTestContext()

	// Components
	btn := NewButton(Style{Color: tcell.ColorWhite}, "Click Me", nil)
	cb := NewCheckbox(ctx, Style{Color: tcell.ColorWhite}, "Check Me", true, nil)
	rb := NewRadioButton(ctx, Style{Color: tcell.ColorWhite}, "Radio Option", "val", true, nil)
	pb := NewProgressBar(Style{Color: tcell.ColorWhite, Width: 20}, 0.5)
	sp := NewSpinner(ctx, Style{Color: tcell.ColorWhite})
	ti := NewTextInput(ctx, Style{Color: tcell.ColorWhite, Width: 20}, "Initial Text", nil)

	// Table
	headers := []string{"ID", "Name", "Status"}
	rows := [][]string{
		{"1", "Alice", "Active"},
		{"2", "Bob", "Pending"},
	}
	tbl := NewTable(Style{Color: tcell.ColorWhite}, headers, rows)

	// Tabs
	tabs := []Tab{
		{Label: "Tab 1", Content: NewText(Style{Color: tcell.ColorWhite}, "Content 1")},
		{Label: "Tab 2", Content: NewText(Style{Color: tcell.ColorWhite}, "Content 2")},
	}
	tbs := NewTabs(ctx, Style{Color: tcell.ColorWhite, ID: "tabs"}, tabs)

	root := NewBox(
		Style{Width: 80, Height: 40, FlexDirection: "column", Padding: Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
		NewText(Style{Color: tcell.ColorGreen}, "--- Form Controls ---"),
		btn,
		cb,
		rb,
		ti,
		NewText(Style{Color: tcell.ColorGreen}, "--- Feedback Controls ---"),
		pb,
		sp,
		NewText(Style{Color: tcell.ColorGreen}, "--- Complex Controls ---"),
		tbl,
		tbs,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 80, MaxH: 40})
	grid := NewGrid(80, 40)

	Render(grid, layout, "", nil)

	verifyVisual(t, grid, "gallery")
}

func TestGenerateTitledBoxVisual(t *testing.T) {
	// 1. Box with title
	box1 := NewBox(Style{Width: 20, Height: 5, Border: true, Title: "Box 1", Color: tcell.ColorWhite},
		NewText(Style{Color: tcell.ColorWhite}, "Content 1"),
	)

	// 2. Box with title that is too long (should not show)
	box2 := NewBox(Style{Width: 10, Height: 5, Border: true, Title: "Too Long Title", Color: tcell.ColorWhite},
		NewText(Style{Color: tcell.ColorWhite}, "Content 2"),
	)

	// 3. Box with title that fits exactly
	box3 := NewBox(Style{Width: 15, Height: 5, Border: true, Title: "Exact Fit", Color: tcell.ColorWhite},
		NewText(Style{Color: tcell.ColorWhite}, "Content 3"),
	)

	root := NewBox(
		Style{Width: 60, Height: 20, FlexDirection: "column", JustifyContent: "space-around"},
		box1,
		box2,
		box3,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 60, MaxH: 20})
	grid := NewGrid(60, 20)

	Render(grid, layout, "", nil)

	verifyVisual(t, grid, "titled_boxes")
}

func TestGenerateFocusVisual(t *testing.T) {
	ctx := makeTestContext()

	btn := NewButton(Style{ID: "btn", Width: 15, Color: tcell.ColorWhite}, "Button", nil)
	ti := NewTextInput(ctx, Style{ID: "ti", Width: 20, Color: tcell.ColorWhite}, "Text Input", nil)

	root := NewBox(
		Style{Width: 50, Height: 10, FlexDirection: "column", JustifyContent: "space-around"},
		btn,
		ti,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 50, MaxH: 10})
	grid := NewGrid(50, 10)

	// Focus the button
	Render(grid, layout, "btn", nil)

	verifyVisual(t, grid, "focus_state")
}

func TestGenerateScrollViewVisual(t *testing.T) {
	ctx := makeTestContext()

	// Child is larger than ScrollView
	child := NewBox(
		Style{Width: 40, Height: 20, Color: tcell.ColorWhite, FlexDirection: "column"},
		NewText(Style{Color: tcell.ColorWhite}, "Line 1"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 2"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 3"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 4"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 5"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 6"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 7"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 8"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 9"),
		NewText(Style{Color: tcell.ColorWhite}, "Line 10"),
	)

	sv := NewScrollView(ctx, Style{Width: 20, Height: 10, Border: true, Title: "Scroll", Color: tcell.ColorWhite}, child)

	root := NewBox(
		Style{Width: 30, Height: 15, JustifyContent: "center"},
		sv,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 30, MaxH: 15})
	grid := NewGrid(30, 15)

	Render(grid, layout, "", nil)

	verifyVisual(t, grid, "scroll_view")
}

func TestGenerateMenuBarVisual(t *testing.T) {
	menus := []Menu{
		{Title: "File", Items: []MenuItem{{Label: "New"}, {Label: "Open"}}},
		{Title: "Edit", Items: []MenuItem{{Label: "Cut"}, {Label: "Copy"}}},
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(50, 8)

	app := NewAppWithScreen(screen)
	app.componentStates["menubar"] = &MenuBarState{OpenMenuIndex: 0}

	renderFn := func(ctx *RenderContext) Node {
		mb := NewMenuBar(ctx, Style{ID: "menubar", Width: 50, Color: tcell.ColorWhite, Background: tcell.ColorBlue}, menus)
		return NewBox(
			Style{Width: 50, Height: 8, FlexDirection: "column"},
			mb,
			NewText(Style{Color: tcell.ColorWhite}, "Content below menu"),
		)
	}

	grid, _, _, _, err := app.RenderFrame(renderFn)
	if err != nil {
		t.Fatal(err)
	}

	verifyVisual(t, grid, "menu_bar")
}

func TestGenerateTabsOverflowVisual(t *testing.T) {
	// Tabs with a fixed narrow width so the header overflows and shows < / > arrows.
	// Scenario 1: scrollOffset=0 (left edge, only '>' shown)
	// Scenario 2: scrollOffset=2 (middle, both '<' and '>' shown)
	// Scenario 3: scrollOffset=4 (right edge, only '<' shown)
	ctx := makeTestContext()

	makeTabs := func(id string) *Tabs {
		return NewTabs(ctx, Style{
			ID:        id,
			Width:     24,
			Color:     tcell.ColorWhite,
			Focusable: true,
		}, []Tab{
			{Label: "Alpha", Content: NewText(Style{Color: tcell.ColorWhite}, "Alpha content")},
			{Label: "Beta", Content: NewText(Style{Color: tcell.ColorWhite}, "Beta content")},
			{Label: "Gamma", Content: NewText(Style{Color: tcell.ColorWhite}, "Gamma content")},
			{Label: "Delta", Content: NewText(Style{Color: tcell.ColorWhite}, "Delta content")},
			{Label: "Epsilon", Content: NewText(Style{Color: tcell.ColorWhite}, "Epsilon content")},
		})
	}

	tabs1 := makeTabs("otabs1") // scroll=0, active=0
	tabs2 := makeTabs("otabs2") // scroll=2, active=2
	tabs3 := makeTabs("otabs3") // scroll=3, active=4

	root := NewBox(
		Style{Width: 30, Height: 18, FlexDirection: "column", Padding: Padding{Top: 1, Left: 1}},
		NewText(Style{Color: tcell.ColorGray}, "Scroll=0, Active=0:"),
		tabs1,
		NewText(Style{Color: tcell.ColorGray}, "Scroll=2, Active=2:"),
		tabs2,
		NewText(Style{Color: tcell.ColorGray}, "Scroll=3, Active=4:"),
		tabs3,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 30, MaxH: 18})
	grid := NewGrid(30, 18)

	compStates := map[string]any{
		"otabs1": &TabsState{ActiveTab: 0, ScrollOffset: 0},
		"otabs2": &TabsState{ActiveTab: 2, ScrollOffset: 2},
		"otabs3": &TabsState{ActiveTab: 4, ScrollOffset: 3},
	}

	Render(grid, layout, "", compStates)

	verifyVisual(t, grid, "tabs_overflow")
}

func TestGenerateTabsVisual(t *testing.T) {
	// Three scenarios rendered side-by-side:
	//  1. First tab active, not focused
	//  2. Second tab active, not focused
	//  3. First tab active, focused
	ctx := makeTestContext()

	makeTabs := func(id string) *Tabs {
		return NewTabs(ctx, Style{
			ID:        id,
			Color:     tcell.ColorWhite,
			Focusable: true,
		}, []Tab{
			{Label: "Home", Content: NewText(Style{Color: tcell.ColorWhite}, "Home content")},
			{Label: "Settings", Content: NewText(Style{Color: tcell.ColorWhite}, "Settings content")},
		})
	}

	tabs1 := makeTabs("tabs1") // tab 0 active, unfocused
	tabs2 := makeTabs("tabs2") // tab 1 active, unfocused
	tabs3 := makeTabs("tabs3") // tab 0 active, focused

	root := NewBox(
		Style{Width: 60, Height: 15, FlexDirection: "column", Padding: Padding{Top: 1, Left: 1}},
		NewText(Style{Color: tcell.ColorGray}, "Tab 0 active (unfocused):"),
		tabs1,
		NewText(Style{Color: tcell.ColorGray}, "Tab 1 active (unfocused):"),
		tabs2,
		NewText(Style{Color: tcell.ColorGray}, "Tab 0 active (focused):"),
		tabs3,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 60, MaxH: 15})
	grid := NewGrid(60, 15)

	compStates := map[string]any{
		"tabs1": &TabsState{ActiveTab: 0},
		"tabs2": &TabsState{ActiveTab: 1},
		"tabs3": &TabsState{ActiveTab: 0},
	}

	Render(grid, layout, "tabs3", compStates)

	verifyVisual(t, grid, "tabs")
}

func TestGenerateModalVisual(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	err := screen.Init()
	if err != nil {
		t.Fatal(err)
	}
	screen.SetSize(50, 20)

	app := NewAppWithScreen(screen)

	renderFn := func(ctx *RenderContext) Node {
		modal := NewModal(ctx, Style{Width: 30, Height: 10, Color: tcell.ColorWhite, Background: tcell.ColorDarkBlue}, 
			NewText(Style{Color: tcell.ColorWhite}, "Modal Content"),
			true, // isOpen
		)

		return NewBox(
			Style{Width: 50, Height: 20, JustifyContent: "center"},
			NewText(Style{Color: tcell.ColorWhite}, "Background Content"),
			modal,
		)
	}

	grid, _, _, _, err := app.RenderFrame(renderFn)
	if err != nil {
		t.Fatal(err)
	}

	verifyVisual(t, grid, "modal_overlay")
}
