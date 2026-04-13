package tz

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// mockMouseEvent implements MouseEvent and tcell.Event for tests.
type mockMouseEvent struct {
	x, y    int
	buttons tcell.ButtonMask
}

func (m mockMouseEvent) Position() (int, int)      { return m.x, m.y }
func (m mockMouseEvent) Buttons() tcell.ButtonMask { return m.buttons }
func (m mockMouseEvent) When() time.Time           { return time.Time{} }

func TestRenderTextInputScrolling(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 5, ID: "input1"}, "1234567890", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 2)

	// Set state with scrollOffset = 5
	states := map[string]any{
		"input1": &TextInputState{cursorOffset: 5, scrollOffset: 5},
	}

	renderToScreen(s, layout, "input1", states)
	s.Show()

	// Should show "67890"
	expected := "67890"
	for i, c := range expected {
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("At col %d, expected '%c', got '%s'", i, c, str)
		}
	}
}

func TestRenderTextInputMultiline(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Multiline: true}, "abc\ndef", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	// Check line 0 "abc"
	expectedLine0 := "abc"
	for i, c := range expectedLine0 {
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("At line 0, col %d, expected '%c', got '%s'", i, c, str)
		}
	}

	// Check line 1 "def"
	expectedLine1 := "def"
	for i, c := range expectedLine1 {
		str, _, _ := s.Get(i, 1)
		if str != string(c) {
			t.Errorf("At line 1, col %d, expected '%c', got '%s'", i, c, str)
		}
	}
}

func TestTextInputHandleEvent(t *testing.T) {
	input := &TextInput{
		Value: "hello",
		Style: Style{ID: "myinput"},
	}
	state := &TextInputState{cursorOffset: 5}

	// Simulate KeyLeft
	ev := tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
	ctx := EventContext{Layout: LayoutResult{H: 1}}

	handled := input.HandleEvent(ev, state, ctx)
	if !handled {
		t.Errorf("Expected event to be handled")
	}
	if state.cursorOffset != 4 {
		t.Errorf("Expected cursorOffset 4, got %d", state.cursorOffset)
	}

	// Simulate typing a rune
	ev = tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone)
	handled = input.HandleEvent(ev, state, ctx)
	if !handled {
		t.Errorf("Expected event to be handled")
	}
	if input.Value != "hell!o" {
		t.Errorf("Expected Value 'hell!o', got '%s'", input.Value)
	}
	if state.cursorOffset != 5 {
		t.Errorf("Expected cursorOffset 5, got %d", state.cursorOffset)
	}
}

func TestOffsetToLineCol(t *testing.T) {
	text := "abc\ndef\nghi"
	tests := []struct {
		offset     int
		wantLine   int
		wantCol    int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{3, 0, 3},
		{4, 1, 0},
		{5, 1, 1},
		{7, 1, 3},
		{8, 2, 0},
		{9, 2, 1},
		{11, 2, 3},
		{12, 2, 3}, // beyond end — clamps to last position
	}
	for _, tc := range tests {
		l, c := offsetToLineCol(text, tc.offset)
		if l != tc.wantLine || c != tc.wantCol {
			t.Errorf("offsetToLineCol(%d): got (%d,%d), want (%d,%d)", tc.offset, l, c, tc.wantLine, tc.wantCol)
		}
	}
}

func TestLineColToOffset(t *testing.T) {
	text := "abc\ndef\nghi"
	tests := []struct {
		line       int
		col        int
		wantOffset int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 3, 3},
		{0, 5, 3},  // col clamped to line length
		{1, 0, 4},
		{1, 1, 5},
		{1, 3, 7},
		{2, 0, 8},
		{2, 3, 11},
		{3, 0, 11}, // line beyond end — clamps to len(text)
	}
	for _, tc := range tests {
		off := lineColToOffset(text, tc.line, tc.col)
		if off != tc.wantOffset {
			t.Errorf("lineColToOffset(%d,%d): got %d, want %d", tc.line, tc.col, off, tc.wantOffset)
		}
	}
}

// --- new feature tests ---

func TestRenderTextInputPlaceholder(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 3)

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 10, ID: "inp"}, "", nil)
	input.Placeholder = "hint text"
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	renderToScreen(s, layout, "", nil)
	s.Show()

	for i, c := range "hint text" {
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("placeholder col %d: got %q, want %q", i, str, string(c))
		}
	}
}

func TestRenderTextInputPlaceholderHiddenWhenNotEmpty(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 3)

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 10, ID: "inp"}, "hi", nil)
	input.Placeholder = "hint text"
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	renderToScreen(s, layout, "", nil)
	s.Show()

	// Should show "hi", not "hint text"
	str0, _, _ := s.Get(0, 0)
	str1, _, _ := s.Get(1, 0)
	if str0 != "h" || str1 != "i" {
		t.Errorf("non-empty input should show value not placeholder, got %q%q", str0, str1)
	}
}

func TestRenderTextInputMask(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 3)

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 10, ID: "inp"}, "secret", nil)
	input.Mask = '*'
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	renderToScreen(s, layout, "", nil)
	s.Show()

	for i := 0; i < 6; i++ {
		str, _, _ := s.Get(i, 0)
		if str != "*" {
			t.Errorf("mask col %d: got %q, want '*'", i, str)
		}
	}
	// Position 6 should be blank (value is only 6 chars wide)
	str, _, _ := s.Get(6, 0)
	if str == "s" || str == "e" {
		t.Errorf("unmasked character leaked at col 6: got %q", str)
	}
}

func TestTextInputDisabled(t *testing.T) {
	input := &TextInput{
		Value:    "hello",
		Disabled: true,
		Style:    Style{ID: "inp", Focusable: true},
	}

	if input.IsFocusable() {
		t.Error("disabled input should not be focusable")
	}

	state := &TextInputState{cursorOffset: 5}
	ctx := EventContext{Layout: LayoutResult{W: 20, H: 1}}

	ev := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	if input.HandleEvent(ev, state, ctx) {
		t.Error("disabled input should not handle key events")
	}
	if input.Value != "hello" {
		t.Errorf("disabled input value changed, got %q", input.Value)
	}

	mev := mockMouseEvent{x: 2, y: 0, buttons: tcell.Button1}
	if input.HandleEvent(mev, state, ctx) {
		t.Error("disabled input should not handle mouse events")
	}
}

func TestTextInputMaxLen(t *testing.T) {
	input := &TextInput{
		Value:  "abc",
		MaxLen: 3,
		Style:  Style{ID: "inp"},
	}
	state := &TextInputState{cursorOffset: 3}
	ctx := EventContext{Layout: LayoutResult{W: 20, H: 1}}

	// Inserting beyond MaxLen should be rejected
	ev := tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone)
	if input.HandleEvent(ev, state, ctx) {
		t.Error("should not accept rune beyond MaxLen")
	}
	if input.Value != "abc" {
		t.Errorf("value should be unchanged, got %q", input.Value)
	}

	// Deletion still works when at MaxLen
	ev = tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone)
	if !input.HandleEvent(ev, state, ctx) {
		t.Error("backspace should work when at MaxLen")
	}
	if input.Value != "ab" {
		t.Errorf("expected 'ab' after backspace, got %q", input.Value)
	}

	// Now there is room — insertion should succeed
	ev = tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	if !input.HandleEvent(ev, state, ctx) {
		t.Error("should accept rune when below MaxLen")
	}
	if input.Value != "abz" {
		t.Errorf("expected 'abz', got %q", input.Value)
	}
}

func TestRenderTextInputScrollIndicatorsBorder(t *testing.T) {
	// Width=5, Border=true: layout W=7, H=3
	// contentX=1, contentW=5, contentY=1
	// With scrollOffset=3 on "0123456789" (10 chars):
	//   visible slice "34567" — overflow left AND right
	//   '<' goes on left border cell  (0, 1)
	//   '>' goes on right border cell (6, 1)
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 3)

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 5, Border: true, ID: "inp"}, "0123456789", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	states := map[string]any{
		"inp": &TextInputState{cursorOffset: 7, scrollOffset: 3},
	}

	renderToScreen(s, layout, "inp", states)
	s.Show()

	left, _, _ := s.Get(0, 1)
	if left != "<" {
		t.Errorf("left scroll indicator: got %q, want '<'", left)
	}
	right, _, _ := s.Get(6, 1)
	if right != ">" {
		t.Errorf("right scroll indicator: got %q, want '>'", right)
	}
}

func TestRenderTextInputMultilineScrollIndicatorsBorder(t *testing.T) {
	// Width=10, Border=true, MaxHeight=7, 7 lines
	// layout W=12, H=7 — contentX=1, contentY=1, contentW=10, contentH=5
	// vScrollOffset=1: lines 1-5 visible, line 0 hidden above, line 6 hidden below
	//   '↑' at (contentX+contentW-1, contentY-1) = (10, 0)  [top border]
	//   '↓' at (contentX+contentW-1, contentY+contentH) = (10, 6) [bottom border]
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 9)

	ctx := makeTestContext()
	value := "line1\nline2\nline3\nline4\nline5\nline6\nline7"
	input := NewTextInput(ctx, Style{Width: 10, Border: true, MaxHeight: 7, Multiline: true, ID: "inp"}, value, nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	states := map[string]any{
		"inp": &TextInputState{cursorOffset: 10, vScrollOffset: 1},
	}

	renderToScreen(s, layout, "", states)
	s.Show()

	up, _, _ := s.Get(10, 0)
	if up != "↑" {
		t.Errorf("up indicator: got %q, want '↑'", up)
	}
	down, _, _ := s.Get(10, 6)
	if down != "↓" {
		t.Errorf("down indicator: got %q, want '↓'", down)
	}
}

func TestTextInputWordNavigation(t *testing.T) {
	input := &TextInput{
		Value: "hello world foo",
		Style: Style{ID: "inp"},
	}
	state := &TextInputState{cursorOffset: 15}
	ctx := EventContext{Layout: LayoutResult{W: 50, H: 1}}

	// Ctrl+Left from end → start of "foo" (offset 12)
	ev := tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModCtrl)
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 12 {
		t.Errorf("Ctrl+Left: expected 12 (start of 'foo'), got %d", state.cursorOffset)
	}

	// Ctrl+Left again → start of "world" (offset 6)
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 6 {
		t.Errorf("Ctrl+Left: expected 6 (start of 'world'), got %d", state.cursorOffset)
	}

	// Ctrl+Right from start of "world" → past end of "world" (offset 11)
	ev = tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModCtrl)
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 11 {
		t.Errorf("Ctrl+Right: expected 11 (end of 'world'), got %d", state.cursorOffset)
	}
}

func TestTextInputCtrlW(t *testing.T) {
	input := &TextInput{
		Value: "hello world",
		Style: Style{ID: "inp"},
	}
	state := &TextInputState{cursorOffset: 11}
	ctx := EventContext{Layout: LayoutResult{W: 50, H: 1}}

	ev := tcell.NewEventKey(tcell.KeyCtrlW, 0, tcell.ModNone)
	input.HandleEvent(ev, state, ctx)

	if input.Value != "hello " {
		t.Errorf("Ctrl+W: expected \"hello \", got %q", input.Value)
	}
	if state.cursorOffset != 6 {
		t.Errorf("Ctrl+W: expected cursorOffset 6, got %d", state.cursorOffset)
	}
}

func TestTextInputMultilineLineHomeEnd(t *testing.T) {
	// "abc\ndef\nghi" — offsets: a=0 b=1 c=2 \n=3 d=4 e=5 f=6 \n=7 g=8 h=9 i=10
	input := &TextInput{
		Value: "abc\ndef\nghi",
		Style: Style{ID: "inp", Multiline: true},
	}
	ctx := EventContext{Layout: LayoutResult{W: 20, H: 5}}

	// Home: mid-line → line start
	state := &TextInputState{cursorOffset: 5} // line 1, col 1
	ev := tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 4 {
		t.Errorf("Home from mid-line: expected 4 (line start), got %d", state.cursorOffset)
	}

	// Home: line start → absolute start
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 0 {
		t.Errorf("Home from line start: expected 0 (absolute start), got %d", state.cursorOffset)
	}

	// End: mid-line → line end
	state = &TextInputState{cursorOffset: 5} // line 1, col 1
	ev = tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 7 {
		t.Errorf("End from mid-line: expected 7 (line end), got %d", state.cursorOffset)
	}

	// End: line end → absolute end
	input.HandleEvent(ev, state, ctx)
	if state.cursorOffset != 11 {
		t.Errorf("End from line end: expected 11 (absolute end), got %d", state.cursorOffset)
	}
}

func TestTextInputMouseClickPositionsCursor(t *testing.T) {
	// Border=true, Padding={Left:1} → content starts at x=2, y=1
	// Click at screen (5,1): clickX = 5 - 0(layoutX) - 1(border) - 1(pad) = 3
	// target = scrollOffset(0) + 3 = 3
	input := &TextInput{
		Value: "hello world",
		Style: Style{ID: "inp", Border: true, Padding: Padding{Left: 1}},
	}
	state := &TextInputState{cursorOffset: 0, scrollOffset: 0}
	ctx := EventContext{Layout: LayoutResult{X: 0, Y: 0, W: 15, H: 3}}

	mev := mockMouseEvent{x: 5, y: 1, buttons: tcell.Button1}
	if !input.HandleEvent(mev, state, ctx) {
		t.Error("mouse click should be handled")
	}
	if state.cursorOffset != 3 {
		t.Errorf("mouse click: expected cursorOffset 3, got %d", state.cursorOffset)
	}
}

func TestTextInputMultilineScrollNotCorrupted(t *testing.T) {
	// Regression: HandleEvent was using raw cursorOffset as horizontal scroll for
	// multiline inputs, pushing all lines off-screen when cursorOffset >> line width.
	input := &TextInput{
		Value: "abc\ndef\nghi",
		Style: Style{ID: "inp", Multiline: true, Border: true},
	}
	// Cursor at end of string (offset 11, line 2 col 3). Line width is 3.
	state := &TextInputState{cursorOffset: 11, scrollOffset: 0}
	ctx := EventContext{Layout: LayoutResult{W: 20, H: 7}}

	ev := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	input.HandleEvent(ev, state, ctx)

	// scrollOffset must stay near the column position (3), not jump to cursorOffset (12)
	if state.scrollOffset > 5 {
		t.Errorf("multiline scrollOffset corrupted: got %d, expected <= 5", state.scrollOffset)
	}
}
