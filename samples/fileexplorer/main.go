package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	var currentUpdate func(tcell.Event)

	var realApp *tizzy.App
	var err error

	initialEntries, _ := os.ReadDir(".")
	sort.Slice(initialEntries, func(i, j int) bool {
		if initialEntries[i].IsDir() != initialEntries[j].IsDir() {
			return initialEntries[i].IsDir()
		}
		return initialEntries[i].Name() < initialEntries[j].Name()
	})

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		previewContent, setPreviewContent := tizzy.UseState[string](ctx, "")
		currentDir, setCurrentDirFn := tizzy.UseState[string](ctx, ".")
		pathInput, setPathInput := tizzy.UseState[string](ctx, currentDir)
		selectedFileIdx, setSelectedFileIdx := tizzy.UseState[int](ctx, -1)

		setCurrentDir := func(s string) {
			setCurrentDirFn(s)
			setPreviewContent("")
			setPathInput(s)
		}

		items, setItems := tizzy.UseState[[]os.DirEntry](ctx, initialEntries)

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

		leftList := tizzy.NewList(ctx, tizzy.Style{ID: "list-left", Focusable: true, Border: true, Title: "Folders", FillWidth: true, FillHeight: true, Color: tcell.ColorGray, FocusColor: tcell.ColorYellow}, currentDir, dirItems, -1, func(item any, index int, selected bool, cursor bool) tizzy.Node {
			label := ""
			if s, ok := item.(string); ok {
				label = s
			} else if d, ok := item.(os.DirEntry); ok {
				label = d.Name() + "/"
			}
			return tizzy.NewListItem(label, selected, cursor)
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
		leftList.OnFocus = func(state *tizzy.ListState) {
			state.ScrollOffset = 0
			state.CursorIndex = 0
		}

		middleListKey := currentDir
		if selectedFileIdx >= 0 && selectedFileIdx < len(fileItems) {
			middleListKey += ":" + fileItems[selectedFileIdx].(os.DirEntry).Name()
		}
		middleList := tizzy.NewList(ctx, tizzy.Style{ID: "list-middle", Focusable: true, Border: true, Title: "Files", FillWidth: true, FillHeight: true, Color: tcell.ColorGray, FocusColor: tcell.ColorYellow}, middleListKey, fileItems, selectedFileIdx, func(item any, index int, selected bool, cursor bool) tizzy.Node {
			label := ""
			if f, ok := item.(os.DirEntry); ok {
				label = f.Name()
			}
			return tizzy.NewListItem(label, selected, cursor)
		}, func(idx int) {
			updatePreview(idx)
		})
		// Preview updates on selection (Enter or click) now

		currentUpdate = func(ev tcell.Event) {
			if key, ok := ev.(*tcell.EventKey); ok {
				if key.Key() == tcell.KeyBackspace || key.Key() == tcell.KeyBackspace2 {
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

		pathInputNode := tizzy.NewTextInput(ctx, tizzy.Style{ID: "path-input", Focusable: true, Border: true, Title: "Path", FillWidth: true, GridRow: 0, GridCol: 0}, pathInput, func(val string) {
			setPathInput(val)
		})
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

		return tizzy.NewGridBox(
			tizzy.Style{Border: true, FillWidth: true, FillHeight: true},
			[]tizzy.GridTrack{tizzy.Flex(1)},
			[]tizzy.GridTrack{tizzy.Fixed(3), tizzy.Flex(1)},

			pathInputNode,

			tizzy.NewGridBox(
				tizzy.Style{GridRow: 1, GridCol: 0, FillWidth: true, FillHeight: true},
				[]tizzy.GridTrack{tizzy.Fixed(25), tizzy.Flex(1), tizzy.Flex(1)},
				[]tizzy.GridTrack{tizzy.Flex(1)},

				// Left Panel
				tizzy.NewBox(tizzy.Style{GridRow: 0, GridCol: 0, FillHeight: true, FillWidth: true},
					leftList,
				),

				// Middle Panel
				tizzy.NewBox(tizzy.Style{GridRow: 0, GridCol: 1, FillHeight: true, FillWidth: true},
					middleList,
				),

				// Right Panel
				tizzy.NewBox(tizzy.Style{GridRow: 0, GridCol: 2, Border: true, FillHeight: true, FillWidth: true, Title: "Preview"},
					tizzy.NewScrollView(ctx, tizzy.Style{ID: "scroll-right", FillHeight: true, FillWidth: true},
						tizzy.NewTextView(tizzy.Style{FillWidth: true}, previewContent),
					),
				),
			),
		)
	}

	realApp, err = tizzy.NewApp()
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
