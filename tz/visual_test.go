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
	ctx := makeTestContext()
	opts := []string{"Option 1", "Option 2", "Option 3"}

	dropdown := NewDropdown(ctx, Style{ID: "dropdown", Width: 20, Color: tcell.ColorWhite, Background: tcell.ColorBlack, Border: true}, opts, 0, nil)

	root := NewBox(
		Style{Width: 40, Height: 15, JustifyContent: "center"},
		dropdown,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 40, MaxH: 15})
	grid := NewGrid(40, 15)

	state := &DropdownState{Open: true, FocusedIndex: 1}
	compStates := map[string]any{"dropdown": state}

	Render(grid, layout, "dropdown", compStates)

	// Replicate overlay rendering from app.go
	res := findLayoutResultByID(layout, "dropdown")
	if res != nil {
		listY := res.Y + res.H
		listW := res.W
		maxH := 3
		listH := maxH

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		popupH := listH + 2

		// Draw Shadow (right and bottom edges only)
		for i := 1; i <= popupH; i++ {
			if listY+i < 15 && res.X+listW < 40 {
				currentCell := grid.Cells[listY+i][res.X+listW]
				grid.SetContent(res.X+listW, listY+i, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
			}
		}
		for j := 1; j <= listW; j++ {
			if listY+popupH < 15 && res.X+j < 40 {
				currentCell := grid.Cells[listY+popupH][res.X+j]
				grid.SetContent(res.X+j, listY+popupH, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
			}
		}

		// Fill background
		for y := 0; y < popupH; y++ {
			for x := 0; x < listW; x++ {
				grid.SetContent(res.X+x, listY+y, ' ', style)
			}
		}

		// Draw Border
		drawBorder(grid, res.X, listY, listW, popupH, "", style)

		// Draw Options
		for i := 0; i < listH; i++ {
			opt := opts[i]
			optStyle := style
			if i == state.FocusedIndex {
				optStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
			}
			curX := res.X + 1
			for _, r := range opt {
				grid.SetContent(curX, listY+1+i, r, optStyle)
				curX++
			}
		}
	}

	verifyVisual(t, grid, "dropdown_layout")
}

func TestGenerateDropdownOpenAboveVisual(t *testing.T) {
	ctx := makeTestContext()
	opts := []string{"Option 1", "Option 2", "Option 3"}

	dropdown := NewDropdown(ctx, Style{ID: "dropdown", Width: 20, Color: tcell.ColorWhite, Background: tcell.ColorBlack, Border: true}, opts, 0, nil)

	// Place dropdown at bottom by putting a spacer box above it
	root := NewBox(
		Style{Width: 40, Height: 15, FlexDirection: "column"},
		NewBox(Style{Height: 10}), // Spacer
		dropdown,
	)

	layout := Layout(root, 0, 0, Constraints{MaxW: 40, MaxH: 15})
	grid := NewGrid(40, 15)

	state := &DropdownState{Open: true, FocusedIndex: 1}
	compStates := map[string]any{"dropdown": state}

	Render(grid, layout, "dropdown", compStates)

	// Replicate overlay rendering from app.go with OpenAbove logic
	res := findLayoutResultByID(layout, "dropdown")
	if res != nil {
		listY := res.Y + res.H
		listW := res.W
		maxH := 3
		listH := maxH

		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
		popupH := listH + 2

		h := 15 // Total grid height
		spaceBelow := h - (res.Y + res.H)
		if spaceBelow < popupH && res.Y >= popupH {
			state.OpenAbove = true
		} else {
			state.OpenAbove = false
		}

		if state.OpenAbove {
			listY = res.Y - popupH
		}

		// Draw Shadow (right and bottom edges only)
		for i := 1; i <= popupH; i++ {
			if listY+i < 15 && res.X+listW < 40 {
				currentCell := grid.Cells[listY+i][res.X+listW]
				grid.SetContent(res.X+listW, listY+i, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
			}
		}
		for j := 1; j <= listW; j++ {
			if listY+popupH < 15 && res.X+j < 40 {
				currentCell := grid.Cells[listY+popupH][res.X+j]
				grid.SetContent(res.X+j, listY+popupH, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
			}
		}

		// Fill background
		for y := 0; y < popupH; y++ {
			for x := 0; x < listW; x++ {
				grid.SetContent(res.X+x, listY+y, ' ', style)
			}
		}

		// Draw Border
		drawBorder(grid, res.X, listY, listW, popupH, "", style)

		// Draw Options
		for i := 0; i < listH; i++ {
			opt := opts[i]
			optStyle := style
			if i == state.FocusedIndex {
				optStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
			}
			curX := res.X + 1
			for _, r := range opt {
				grid.SetContent(curX, listY+1+i, r, optStyle)
				curX++
			}
		}
	}

	verifyVisual(t, grid, "dropdown_above_layout")
}
