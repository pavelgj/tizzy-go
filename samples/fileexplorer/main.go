package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func getCompletions(path string) []string {
	var dir string
	var base string

	if strings.HasSuffix(path, "/") || path == "/" {
		dir = path
		base = ""
	} else {
		dir = filepath.Dir(path)
		base = filepath.Base(path)
		if dir == path {
			dir = "."
			base = path
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var sugs []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, base) {
			sugs = append(sugs, name)
		}
	}
	return sugs
}

func main() {
	var currentUpdate func(tcell.Event)

	var realApp *tz.App
	var err error

	initialEntries, _ := os.ReadDir(".")
	sort.Slice(initialEntries, func(i, j int) bool {
		if initialEntries[i].IsDir() != initialEntries[j].IsDir() {
			return initialEntries[i].IsDir()
		}
		return initialEntries[i].Name() < initialEntries[j].Name()
	})

	render := func(ctx *tz.RenderContext) tz.Node {
		previewContent, setPreviewContent := tz.UseState[string](ctx, "")
		currentDir, setCurrentDirFn := tz.UseState[string](ctx, ".")
		pathInput, setPathInput := tz.UseState[string](ctx, currentDir)
		selectedFileIdx, setSelectedFileIdx := tz.UseState[int](ctx, -1)
		popupOpen, setPopupOpen := tz.UseState[bool](ctx, false)
		filteredSuggestions, setFilteredSuggestions := tz.UseState[[]string](ctx, nil)
		selectedSug, setSelectedSug := tz.UseState[int](ctx, 0)
		cursorOverride, setCursorOverride := tz.UseState[*int](ctx, nil)

		setCurrentDir := func(s string) {
			setCurrentDirFn(s)
			setPreviewContent("")
			setPathInput(s)
		}

		items, setItems := tz.UseState[[]os.DirEntry](ctx, initialEntries)

		var dirs []os.DirEntry
		var files []os.DirEntry

		for _, item := range items {
			if item.IsDir() {
				dirs = append(dirs, item)
			} else {
				files = append(files, item)
			}
		}

		updatePreview := func(idx int) {
			if idx < len(files) {
				entry := files[idx]
				if !entry.IsDir() {
					f, err := os.Open(filepath.Join(currentDir, entry.Name()))
					if err == nil {
						defer func() { _ = f.Close() }()
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

		leftList := tz.NewList(ctx, tz.Style{ID: "list-left", Focusable: true, Border: true, Title: "Folders", FillWidth: true, FillHeight: true, Color: tcell.ColorGray, FocusColor: tcell.ColorYellow}, currentDir, dirItems, -1, func(item any, index int, selected bool, cursor bool) tz.Node {
			label := ""
			if s, ok := item.(string); ok {
				label = s
			} else if d, ok := item.(os.DirEntry); ok {
				label = d.Name() + "/"
			}
			return tz.NewListItem(label, selected, cursor)
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
		leftList.OnFocus = func(state *tz.ListState) {
			state.ScrollOffset = 0
			state.CursorIndex = 0
		}

		middleListKey := currentDir
		if selectedFileIdx >= 0 && selectedFileIdx < len(fileItems) {
			middleListKey += ":" + fileItems[selectedFileIdx].(os.DirEntry).Name()
		}
		middleList := tz.NewList(ctx, tz.Style{ID: "list-middle", Focusable: true, Border: true, Title: "Files", FillWidth: true, FillHeight: true, Color: tcell.ColorGray, FocusColor: tcell.ColorYellow}, middleListKey, fileItems, selectedFileIdx, func(item any, index int, selected bool, cursor bool) tz.Node {
			label := ""
			if f, ok := item.(os.DirEntry); ok {
				label = f.Name()
			}
			return tz.NewListItem(label, selected, cursor)
		}, func(idx int) {
			updatePreview(idx)
			if idx >= 0 && idx < len(files) {
				setPathInput(filepath.Join(currentDir, files[idx].Name()))
			}
		})
		// Preview updates on selection (Enter or click) now

		currentUpdate = func(ev tcell.Event) {
			if key, ok := ev.(*tcell.EventKey); ok {
				if popupOpen {
					if key.Key() == tcell.KeyDown {
						setSelectedSug((selectedSug + 1) % len(filteredSuggestions))
						return
					} else if key.Key() == tcell.KeyUp {
						setSelectedSug((selectedSug - 1 + len(filteredSuggestions)) % len(filteredSuggestions))
						return
					} else if key.Key() == tcell.KeyEnter {
						if selectedSug >= 0 && selectedSug < len(filteredSuggestions) {
							sug := filteredSuggestions[selectedSug]
							lastSlash := strings.LastIndex(pathInput, "/")
							newVal := pathInput[:lastSlash+1] + sug
							setPathInput(newVal)
							setPopupOpen(false)
							newOffset := len(newVal)
							setCursorOverride(&newOffset)
						}
						return
					} else if key.Key() == tcell.KeyEscape {
						setPopupOpen(false)
						return
					}
				}
				if key.Key() == tcell.KeyBackspace || key.Key() == tcell.KeyBackspace2 {
					if ctx.GetFocusedID() == "path-input" {
						return
					}
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

		pathInputNode := tz.NewTextInput(ctx, tz.Style{ID: "path-input", Focusable: true, Border: true, Title: "Path", FillWidth: true, GridRow: 0, GridCol: 0}, pathInput, func(val string) {
			setPathInput(val)

			if strings.HasSuffix(val, "/") || val == "/" {
				sugs := getCompletions(val)
				if len(sugs) > 0 {
					setFilteredSuggestions(sugs)
					setPopupOpen(true)
					setSelectedSug(0)
				} else {
					setPopupOpen(false)
				}
			} else if popupOpen {
				sugs := getCompletions(val)
				if len(sugs) > 0 {
					setFilteredSuggestions(sugs)
				} else {
					setPopupOpen(false)
				}
			}
		})
		pathInputNode.Cursor = cursorOverride
		if cursorOverride != nil {
			setCursorOverride(nil)
		}
		pathInputNode.OnSubmit = func(val string) {
			info, err := os.Stat(val)
			if err == nil {
				if info.IsDir() {
					setCurrentDirFn(val)
					entries, err := os.ReadDir(val)
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
					setCurrentDirFn(filepath.Dir(val))
					f, err := os.Open(val)
					if err == nil {
						defer func() { _ = f.Close() }()
						buf := make([]byte, 1000)
						n, _ := f.Read(buf)
						setPreviewContent(string(buf[:n]))
					} else {
						setPreviewContent("Error reading file: " + err.Error())
					}
					entries, err := os.ReadDir(filepath.Dir(val))
					if err == nil {
						sort.Slice(entries, func(i, j int) bool {
							if entries[i].IsDir() != entries[j].IsDir() {
								return entries[i].IsDir()
							}
							return entries[i].Name() < entries[j].Name()
						})
						setItems(entries)

						// Find index of the file to select it
						var files []os.DirEntry
						for _, item := range entries {
							if !item.IsDir() {
								files = append(files, item)
							}
						}
						fileName := filepath.Base(val)
						for i, f := range files {
							if f.Name() == fileName {
								setSelectedFileIdx(i)
								break
							}
						}
					}
				}
			}
		}

		listItems := []tz.Node{}
		for i, sug := range filteredSuggestions {
			style := tz.Style{Padding: tz.Padding{Left: 1, Right: 1}}
			if i == selectedSug {
				style.Background = tcell.ColorYellow
				style.Color = tcell.ColorBlack
			}
			listItems = append(listItems, tz.NewText(style, sug))
		}

		popupNode := tz.NewPopup(
			ctx,
			tz.Style{
				Border:     true,
				Background: tcell.ColorGray,
				Width:      30,
			},
			tz.NewBox(
				tz.Style{FlexDirection: "column"},
				listItems...,
			),
			10, // X
			3,  // Y (below path input)
			popupOpen && len(filteredSuggestions) > 0,
		)

		return tz.NewGridBox(
			tz.Style{Border: true, FillWidth: true, FillHeight: true},
			[]tz.GridTrack{tz.Flex(1)},
			[]tz.GridTrack{tz.Fixed(3), tz.Flex(1)},
			pathInputNode,
			popupNode,

			tz.NewGridBox(
				tz.Style{GridRow: 1, GridCol: 0, FillWidth: true, FillHeight: true},
				[]tz.GridTrack{tz.Fixed(25), tz.Flex(1), tz.Flex(1)},
				[]tz.GridTrack{tz.Flex(1)},

				// Left Panel
				tz.NewBox(tz.Style{GridRow: 0, GridCol: 0, FillHeight: true, FillWidth: true},
					leftList,
				),

				// Middle Panel
				tz.NewBox(tz.Style{GridRow: 0, GridCol: 1, FillHeight: true, FillWidth: true},
					middleList,
				),

				// Right Panel
				tz.NewBox(tz.Style{GridRow: 0, GridCol: 2, Border: true, FillHeight: true, FillWidth: true, Title: "Preview"},
					tz.NewScrollView(ctx, tz.Style{ID: "scroll-right", FillHeight: true, FillWidth: true},
						tz.NewTextView(tz.Style{FillWidth: true}, previewContent),
					),
				),
			),
		)
	}

	realApp, err = tz.NewApp()
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
		if currentUpdate != nil {
			currentUpdate(ev)
		}
	}

	if err := realApp.Run(render, wrappedUpdate); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
