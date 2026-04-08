package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gdamore/tcell/v2"
	"splotch/splotch"
)

func updatePreview(currentDir string, items []os.DirEntry, idx int, activePanel string, setPreviewContent func(string)) {
	if activePanel == "middle" && idx < len(items) {
		entry := items[idx]
		if !entry.IsDir() {
			f, err := os.Open(filepath.Join(currentDir, entry.Name()))
			if err == nil {
				defer f.Close()
				buf := make([]byte, 1000)
				n, _ := f.Read(buf)
				setPreviewContent(string(buf[:n]))
			} else {
				setPreviewContent("Error reading file: " + err.Error())
			}
		} else {
			setPreviewContent("[Directory]")
		}
	}
}

func main() {
	var currentDir string
	var items []os.DirEntry
	var previewContent string

	var setCurrentDir func(string)
	var setItems func([]os.DirEntry)
	var setPreviewContent func(string)

	var realApp *splotch.App
	var err error

	initialEntries, _ := os.ReadDir(".")
	sort.Slice(initialEntries, func(i, j int) bool {
		if initialEntries[i].IsDir() != initialEntries[j].IsDir() {
			return initialEntries[i].IsDir()
		}
		return initialEntries[i].Name() < initialEntries[j].Name()
	})

	render := func(ctx *splotch.RenderContext) splotch.Node {
		previewContentObj, setPreviewContentFn := splotch.UseState[string](ctx, "")
		previewContent = previewContentObj
		setPreviewContent = func(s string) { setPreviewContentFn(s) }

		currentDirObj, setCurrentDirFn := splotch.UseState[string](ctx, ".")
		currentDir = currentDirObj
		setCurrentDir = func(s string) {
			setCurrentDirFn(s)
			setPreviewContentFn("")
		}

		itemsObj, setItemsFn := splotch.UseState[[]os.DirEntry](ctx, initialEntries)
		items = itemsObj
		setItems = func(e []os.DirEntry) { setItemsFn(e) }

		var dirs []os.DirEntry
		var files []os.DirEntry

		for _, item := range items {
			if item.IsDir() {
				dirs = append(dirs, item)
			} else {
				files = append(files, item)
			}
		}

		var dirItems []any
		hasParent := filepath.Dir(currentDir) != currentDir
		if hasParent {
			dirItems = append(dirItems, "..")
		}
		for _, d := range dirs {
			dirItems = append(dirItems, d)
		}

		var fileItems []any
		for _, f := range files {
			fileItems = append(fileItems, f)
		}

		leftList := splotch.NewList(ctx, splotch.Style{ID: "list-left", Focusable: true}, currentDir, dirItems, func(item any, index int, selected bool, cursor bool) splotch.Node {
			label := ""
			if s, ok := item.(string); ok {
				label = s
			} else if d, ok := item.(os.DirEntry); ok {
				label = d.Name() + "/"
			}
			return splotch.NewListItem(label, selected, cursor)
		}, func(idx int) {
			// OnSelect
			if hasParent && idx == 0 {
				parent := filepath.Dir(currentDir)
				setCurrentDir(parent)
				entries, err := os.ReadDir(parent)
				if err == nil {
					sort.Slice(entries, func(i, j int) bool {
						if entries[i].IsDir() != entries[j].IsDir() {
							return entries[i].IsDir()
						}
						return entries[i].Name() < entries[j].Name()
					})
					setItems(entries)
				}
			} else {
				realIdx := idx
				if hasParent {
					realIdx = idx - 1
				}
				if realIdx >= 0 && realIdx < len(dirs) {
					newDir := filepath.Join(currentDir, dirs[realIdx].Name())
					setCurrentDir(newDir)
					entries, err := os.ReadDir(newDir)
					if err == nil {
						sort.Slice(entries, func(i, j int) bool {
							if entries[i].IsDir() != entries[j].IsDir() {
								return entries[i].IsDir()
							}
							return entries[i].Name() < entries[j].Name()
						})
						setItems(entries)
					}
				}
			}
		})

		middleList := splotch.NewList(ctx, splotch.Style{ID: "list-middle", Focusable: true}, currentDir, fileItems, func(item any, index int, selected bool, cursor bool) splotch.Node {
			label := ""
			if f, ok := item.(os.DirEntry); ok {
				label = f.Name()
			}
			return splotch.NewListItem(label, selected, cursor)
		}, func(idx int) {
			updatePreview(currentDir, files, idx, "middle", setPreviewContent)
		})

		// Preview updates on selection (Enter or click) now

		focused := ctx.GetFocusedID()
		leftBorder := tcell.ColorGray
		if focused == "list-left" {
			leftBorder = tcell.ColorYellow
		}
		middleBorder := tcell.ColorGray
		if focused == "list-middle" {
			middleBorder = tcell.ColorYellow
		}

		return splotch.NewGridBox(
			splotch.Style{Border: true, FillWidth: true, FillHeight: true},
			[]splotch.GridTrack{splotch.Fixed(25), splotch.Flex(1), splotch.Flex(1)},
			[]splotch.GridTrack{splotch.Flex(1)},

			// Left Panel
			splotch.NewBox(splotch.Style{GridRow: 0, GridCol: 0, Border: true, FillHeight: true, FillWidth: true, Color: leftBorder},
				splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Folders"),
				leftList,
			),

			// Middle Panel
			splotch.NewBox(splotch.Style{GridRow: 0, GridCol: 1, Border: true, FillHeight: true, FillWidth: true, Color: middleBorder},
				splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Files"),
				middleList,
			),

			// Right Panel
			splotch.NewBox(splotch.Style{GridRow: 0, GridCol: 2, Border: true, FillHeight: true, FillWidth: true},
				splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Preview"),
				splotch.NewScrollView(ctx, splotch.Style{ID: "scroll-right", FillHeight: true, FillWidth: true},
					splotch.NewTextView(splotch.Style{FillWidth: true}, previewContent),
				),
			),
		)
	}

	update := func(ev tcell.Event) {
		if key, ok := ev.(*tcell.EventKey); ok {
			if key.Key() == tcell.KeyTab {
				focused := realApp.GetFocusedID()
				if focused == "list-left" {
					if stateObj, ok := realApp.GetComponentState("list-middle"); ok {
						state := stateObj.(*splotch.ListState)
						state.ScrollOffset = 0
						state.CursorIndex = 0
					}
				} else {
					if stateObj, ok := realApp.GetComponentState("list-left"); ok {
						state := stateObj.(*splotch.ListState)
						state.ScrollOffset = 0
						state.CursorIndex = 0
					}
				}
			} else if key.Key() == tcell.KeyBackspace || key.Key() == tcell.KeyBackspace2 {
				parent := filepath.Dir(currentDir)
				if parent != currentDir {
					setCurrentDir(parent)
					entries, err := os.ReadDir(parent)
					if err == nil {
						sort.Slice(entries, func(i, j int) bool {
							if entries[i].IsDir() != entries[j].IsDir() {
								return entries[i].IsDir()
							}
							return entries[i].Name() < entries[j].Name()
						})
						setItems(entries)
					}
				}
			}
		}
	}

	realApp, err = splotch.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	wrappedUpdate := func(ev tcell.Event) {
		if key, ok := ev.(*tcell.EventKey); ok {
			if key.Key() == tcell.KeyEscape || key.Key() == tcell.KeyCtrlC {
				realApp.Stop()
				return
			}
		}
		update(ev)
	}

	if err := realApp.Run(render, wrappedUpdate); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
