package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pavelgj/tizzy-go/tz"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(func(ctx *tz.RenderContext) tz.Node {
		return tz.NewBox(
			tz.Style{FlexDirection: "column", FillWidth: true, FillHeight: true},
			header(),
			tz.NewTabs(ctx, tz.Style{
				ID:        "nav",
				Focusable: true,
				FillWidth: true,
				FillHeight: true,
				Color:     tcell.ColorWhite,
				Padding:   tz.Padding{Left: 2},
			}, []tz.Tab{
				{Label: "Spinners", Content: spinnersPage(ctx)},
				{Label: "Progress", Content: progressPage(ctx)},
				{Label: "Transitions", Content: transitionsPage(ctx)},
			}),
		)
	}, nil)

	if err != nil {
		log.Fatal(err)
	}
}

func header() tz.Node {
	return tz.NewBox(
		tz.Style{
			FillWidth: true,
			Padding:   tz.Padding{Top: 1, Bottom: 1, Left: 2},
			Background: tcell.ColorDarkBlue,
		},
		tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Animation Samples"),
		tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Left: 2}},
			"← → to switch tabs when nav is focused • Tab to move focus • Esc to quit"),
	)
}

// ── Spinners page ─────────────────────────────────────────────────────────────
// Shows NewSpinner and NewSpinnerCustom with different frame sets and speeds.

func spinnersPage(ctx *tz.RenderContext) tz.Node {
	s1 := tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorTeal})

	s2 := tz.NewSpinnerCustom(ctx, tz.Style{Color: tcell.ColorGreen},
		[]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		60*time.Millisecond)

	s3 := tz.NewSpinnerCustom(ctx, tz.Style{Color: tcell.ColorYellow},
		[]string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█", "▉", "▊", "▋", "▌", "▍", "▎"},
		50*time.Millisecond)

	s4 := tz.NewSpinnerCustom(ctx, tz.Style{Color: tcell.ColorPurple},
		[]string{".  ", ".. ", "...", " ..", "  .", "   "},
		180*time.Millisecond)

	return tz.NewBox(
		tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 2, Left: 4}},
		tz.NewText(tz.Style{Color: tcell.ColorWhite}, "NewSpinner / NewSpinnerCustom"),
		tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Bottom: 1}},
			"Each spinner drives its own animation slot on the shared scheduler."),

		row(s1, "Default  |/-\\  @ 100ms/frame"),
		row(s2, "Braille  ⠋⠙⠹⠸  @ 60ms/frame"),
		row(s3, "Block    ▏▎▍▌▋  @ 50ms/frame"),
		row(s4, "Dots     .  .. ...  @ 180ms/frame"),
	)
}

// ── Progress page ─────────────────────────────────────────────────────────────
// Shows UseTween tracking a stepped integer target with smooth interpolation.

func progressPage(ctx *tz.RenderContext) tz.Node {
	const barW = 50
	const step = 10

	target, setTarget := tz.UseState(ctx, 40)

	// UseTween animates 'current' smoothly toward 'target' on every change.
	current := tz.UseTween(ctx, target, 350*time.Millisecond, tz.EaseOut)

	filled := current * barW / 100
	if filled < 0 {
		filled = 0
	}
	if filled > barW {
		filled = barW
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)

	return tz.NewBox(
		tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 2, Left: 4}},
		tz.NewText(tz.Style{Color: tcell.ColorWhite}, "UseTween — smooth value tracking"),
		tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Bottom: 1}},
			"Press + / - to step the target; the bar animates smoothly toward it."),

		tz.NewText(tz.Style{Color: tcell.ColorTeal}, bar),
		tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
			fmt.Sprintf("target %3d%%   current %3d%%", target, current)),

		tz.NewBox(
			tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 2}},
			tz.NewButton(tz.Style{ID: "btn-minus", Focusable: true, Border: true}, "  -  ", func() {
				v := target - step
				if v < 0 {
					v = 0
				}
				setTarget(v)
			}),
			tz.NewButton(tz.Style{ID: "btn-plus", Focusable: true, Border: true, Margin: tz.Margin{Left: 1}}, "  +  ", func() {
				v := target + step
				if v > 100 {
					v = 100
				}
				setTarget(v)
			}),
			tz.NewButton(tz.Style{ID: "btn-zero", Focusable: true, Border: true, Margin: tz.Margin{Left: 1}}, " Reset ", func() {
				setTarget(0)
			}),
			tz.NewButton(tz.Style{ID: "btn-full", Focusable: true, Border: true, Margin: tz.Margin{Left: 1}}, " Full ", func() {
				setTarget(100)
			}),
		),

		tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 2}},
			"Mid-flight retargeting: press buttons while the bar is still moving."),
	)
}

// ── Transitions page ──────────────────────────────────────────────────────────
// Shows UseAnimation with WithManualTrigger for a flash effect and UseTween
// for a slide-open panel.

func transitionsPage(ctx *tz.RenderContext) tz.Node {
	return tz.NewBox(
		tz.Style{FlexDirection: "row", Padding: tz.Padding{Top: 2, Left: 4}},
		tz.NewBox(
			tz.Style{FlexDirection: "column", Width: 42, Margin: tz.Margin{Right: 4}},
			flashDemo(ctx),
			slideDemo(ctx),
		),
		easingDemo(ctx),
	)
}

// flashDemo — a button triggers a brief yellow highlight that decays.
func flashDemo(ctx *tz.RenderContext) tz.Node {
	progress, flash := tz.UseAnimation(ctx, 600*time.Millisecond, tz.EaseIn,
		tz.WithManualTrigger())

	// Intensity decays from 255 → 0 as progress goes 0 → 1.
	intensity := int32((1.0 - progress) * 255)
	bg := tcell.NewRGBColor(intensity, intensity, 0)

	return tz.NewBox(
		tz.Style{FlexDirection: "column", Margin: tz.Margin{Bottom: 2}},
		tz.NewText(tz.Style{Color: tcell.ColorWhite}, "UseAnimation + WithManualTrigger — flash on demand"),
		tz.NewBox(
			tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
			tz.NewButton(tz.Style{ID: "btn-flash", Focusable: true, Border: true}, " Flash! ", flash),
			tz.NewBox(
				tz.Style{
					Width:      24,
					Background: bg,
					Margin:     tz.Margin{Left: 2},
					Padding:    tz.Padding{Left: 1},
				},
				tz.NewText(tz.Style{Color: tcell.ColorBlack}, "  flash indicator  "),
			),
		),
	)
}

// slideDemo — a button toggles a panel open/closed via UseTween on its height.
func slideDemo(ctx *tz.RenderContext) tz.Node {
	open, setOpen := tz.UseState(ctx, false)

	targetH := 0
	if open {
		targetH = 4
	}
	h := tz.UseTween(ctx, targetH, 220*time.Millisecond, tz.EaseOut)

	label := " Open panel "
	if open {
		label = " Close panel "
	}

	var panel tz.Node
	if h > 0 {
		panel = tz.NewBox(
			tz.Style{
				Height:     h,
				Width:      38,
				Background: tcell.ColorDarkBlue,
				Padding:    tz.Padding{Left: 2, Top: 1},
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Sliding panel — height via UseTween"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, fmt.Sprintf("h = %d", h)),
		)
	} else {
		panel = tz.NewBox(tz.Style{})
	}

	return tz.NewBox(
		tz.Style{FlexDirection: "column", Margin: tz.Margin{Bottom: 2}},
		tz.NewText(tz.Style{Color: tcell.ColorWhite}, "UseTween — slide open / close"),
		tz.NewButton(tz.Style{ID: "btn-slide", Focusable: true, Border: true, Margin: tz.Margin{Top: 1}},
			label, func() { setOpen(!open) }),
		panel,
	)
}

// easingDemo — three bars animate on mount, each with a different easing curve.
// Press "Replay" to restart them all.
func easingDemo(ctx *tz.RenderContext) tz.Node {
	const maxW = 40

	p1, replay1 := tz.UseAnimation(ctx, 1200*time.Millisecond, tz.Linear, tz.WithManualTrigger())
	p2, replay2 := tz.UseAnimation(ctx, 1200*time.Millisecond, tz.EaseIn, tz.WithManualTrigger())
	p3, replay3 := tz.UseAnimation(ctx, 1200*time.Millisecond, tz.EaseOut, tz.WithManualTrigger())
	p4, replay4 := tz.UseAnimation(ctx, 1200*time.Millisecond, tz.EaseInOut, tz.WithManualTrigger())

	replay := func() {
		replay1()
		replay2()
		replay3()
		replay4()
	}

	// Auto-play once on mount.
	played, setPlayed := tz.UseState(ctx, false)
	if !played {
		replay()
		setPlayed(true)
	}

	bar := func(p float64, label string) tz.Node {
		w := int(p * maxW)
		return tz.NewBox(
			tz.Style{FlexDirection: "column", Margin: tz.Margin{Top: 1}},
			tz.NewText(tz.Style{Color: tcell.ColorGray}, label),
			tz.NewText(tz.Style{Color: tcell.ColorTeal}, strings.Repeat("█", w)),
		)
	}

	return tz.NewBox(
		tz.Style{FlexDirection: "column"},
		tz.NewText(tz.Style{Color: tcell.ColorWhite}, "UseAnimation — easing curves"),
		bar(p1, "Linear"),
		bar(p2, "EaseIn"),
		bar(p3, "EaseOut"),
		bar(p4, "EaseInOut"),
		tz.NewButton(tz.Style{ID: "btn-replay", Focusable: true, Border: true, Margin: tz.Margin{Top: 1}},
			" Replay ", replay),
	)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func row(spinner *tz.Spinner, label string) tz.Node {
	return tz.NewBox(
		tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
		spinner,
		tz.NewText(tz.Style{Margin: tz.Margin{Left: 2}, Color: tcell.ColorGray}, label),
	)
}
