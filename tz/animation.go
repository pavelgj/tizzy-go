package tz

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// ── Easing ───────────────────────────────────────────────────────────────────

// EasingFn maps a normalised time value [0, 1] to a normalised progress value
// [0, 1]. Pass one to UseAnimation or UseTween to control the interpolation
// curve.
type EasingFn func(t float64) float64

// Built-in easing functions. These cover the vast majority of TUI animation
// needs.
var (
	Linear    EasingFn = func(t float64) float64 { return t }
	EaseIn    EasingFn = func(t float64) float64 { return t * t }
	EaseOut   EasingFn = func(t float64) float64 { return t * (2 - t) }
	EaseInOut EasingFn = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}
)

// ── Scheduler ────────────────────────────────────────────────────────────────

// animSlot is a single registered animation. step is called on every tick and
// returns false when the animation is finished, signalling the scheduler to
// remove it automatically.
type animSlot struct {
	step func(now time.Time) bool
}

// animationScheduler drives all active animations from a single goroutine.
// Step functions call state setters which call MarkDirty; MarkDirty posts an
// EventTick to unblock PollEvent so the render loop picks up the changes.
type animationScheduler struct {
	mu    sync.Mutex
	slots map[string]animSlot
}

func newAnimationScheduler() *animationScheduler {
	return &animationScheduler{slots: make(map[string]animSlot)}
}

// register adds or replaces an animation slot. Safe to call concurrently.
func (s *animationScheduler) register(id string, step func(time.Time) bool) {
	s.mu.Lock()
	s.slots[id] = animSlot{step: step}
	s.mu.Unlock()
}

// unregister removes an animation slot. Safe to call concurrently.
func (s *animationScheduler) unregister(id string) {
	s.mu.Lock()
	delete(s.slots, id)
	s.mu.Unlock()
}

// run ticks at fps Hz and steps every registered animation. Slots whose step
// function returns false are removed automatically. Runs until the process
// exits; launch with go.
func (s *animationScheduler) run(fps int) {
	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()
	for now := range ticker.C {
		s.mu.Lock()
		if len(s.slots) == 0 {
			s.mu.Unlock()
			continue
		}
		var done []string
		for id, slot := range s.slots {
			if !slot.step(now) {
				done = append(done, id)
			}
		}
		for _, id := range done {
			delete(s.slots, id)
		}
		s.mu.Unlock()
		// State setters called by step functions invoke MarkDirty, which posts
		// an EventTick to wake up PollEvent. No explicit post needed here.
	}
}

// ── animID ───────────────────────────────────────────────────────────────────

// animID returns a stable, unique key for the scheduler's slots map. It
// encodes the app pointer (unique per process) and the hook index captured
// before any UseState/UseEffect calls inside the hook, so the key is the same
// on every render of the same component.
func animID(ctx *RenderContext, kind string, baseIdx int) string {
	return fmt.Sprintf("%p/%s/%d", ctx.app, kind, baseIdx)
}

// ── UseAnimation ─────────────────────────────────────────────────────────────

// AnimOpt configures UseAnimation behaviour.
type AnimOpt func(*animCfg)

type animCfg struct {
	loop   bool
	manual bool
}

// WithLoop causes the animation to restart from the beginning automatically
// when it reaches 1.0, looping indefinitely.
func WithLoop() AnimOpt { return func(c *animCfg) { c.loop = true } }

// WithManualTrigger causes the animation to remain at 0.0 on mount and start
// only when the returned trigger function is called (e.g. from an event
// handler).
func WithManualTrigger() AnimOpt { return func(c *animCfg) { c.manual = true } }

// animControl holds mutable per-animation state shared between the hook and
// the scheduler step function. The mu field guards concurrent access: the
// scheduler reads/writes animControl while holding scheduler.mu, and trigger
// reads/writes animControl after releasing scheduler.mu (acquired separately
// to avoid lock-order inversion).
type animControl struct {
	mu      sync.Mutex
	start   time.Time
	running bool
}

// UseAnimation returns a progress value in [0.0, 1.0] that advances from 0 to
// 1 over duration according to the easing function, and a trigger function.
//
// By default the animation starts immediately on mount. Use WithManualTrigger
// to defer the start; call the returned trigger function to begin it.
// Use WithLoop to repeat the animation indefinitely.
//
// Example — fade in on mount:
//
//	progress := tz.UseAnimation(ctx, 400*time.Millisecond, tz.EaseOut)
//	v := uint8(progress * 255)
//	color := tcell.NewRGBColor(int32(v), int32(v), int32(v))
//
// Example — flash on keypress:
//
//	progress, flash := tz.UseAnimation(ctx, 200*time.Millisecond, tz.EaseOut,
//	    tz.WithManualTrigger())
//	// call flash() from HandleEvent to start the animation
func UseAnimation(ctx *RenderContext, duration time.Duration, easing EasingFn, opts ...AnimOpt) (float64, func()) {
	cfg := &animCfg{}
	for _, opt := range opts {
		opt(cfg)
	}

	baseIdx := ctx.hookIndex
	id := animID(ctx, "anim", baseIdx)

	progress, setProgress := UseState(ctx, 0.0)
	ctrl, _ := UseState(ctx, &animControl{
		start:   time.Now(),
		running: !cfg.manual,
	})

	sched := ctx.app.scheduler

	// stepFn is defined before trigger so trigger can re-register it.
	var stepFn func(time.Time) bool
	stepFn = func(now time.Time) bool {
		ctrl.mu.Lock()
		defer ctrl.mu.Unlock()
		if !ctrl.running {
			return false
		}
		t := now.Sub(ctrl.start).Seconds() / duration.Seconds()
		if t >= 1.0 {
			if cfg.loop {
				ctrl.start = now
				t = 0
			} else {
				setProgress(easing(1.0))
				ctrl.running = false
				return false
			}
		}
		setProgress(easing(t))
		return true
	}

	trigger := func() {
		ctrl.mu.Lock()
		ctrl.start = time.Now()
		ctrl.running = true
		ctrl.mu.Unlock()
		// Register after releasing ctrl.mu to avoid lock-order inversion
		// (scheduler holds sched.mu then acquires ctrl.mu inside stepFn).
		if sched != nil {
			sched.register(id, stepFn)
		}
	}

	ctx.UseEffect(func() func() {
		if !cfg.manual && sched != nil {
			sched.register(id, stepFn)
		}
		return func() {
			if sched != nil {
				sched.unregister(id)
			}
		}
	})

	return progress, trigger
}

// ── UseTween ─────────────────────────────────────────────────────────────────

// Numeric is the type constraint for values that UseTween can interpolate.
type Numeric interface {
	~int | ~float64
}

// tweenControl holds mutable state shared between a UseTween hook and the
// scheduler. The mu field guards concurrent access (see animControl comment).
type tweenControl struct {
	mu      sync.Mutex
	fromF   float64
	toF     float64
	start   time.Time
	running bool
}

// UseTween smoothly interpolates a numeric value toward target whenever target
// changes, animating over duration using the given easing function. It behaves
// like a CSS transition: the first render shows the value immediately; only
// subsequent changes trigger animation.
//
// When target changes mid-animation, the tween picks up from the current
// interpolated value and animates toward the new target.
//
// Example — animated sidebar width:
//
//	targetW := 0
//	if open {
//	    targetW = 40
//	}
//	w := tz.UseTween(ctx, targetW, 200*time.Millisecond, tz.EaseOut)
//	return tz.NewBox(ctx, tz.Style{Width: w, FillHeight: true}, content)
//
// Example — smooth progress bar fill:
//
//	smoothPct := tz.UseTween(ctx, downloadPercent, 150*time.Millisecond, tz.Linear)
//	filled := int(smoothPct * float64(barWidth))
func UseTween[T Numeric](ctx *RenderContext, target T, duration time.Duration, easing EasingFn) T {
	baseIdx := ctx.hookIndex
	id := animID(ctx, "tween", baseIdx)

	current, setCurrent := UseState(ctx, target)
	ctrl, _ := UseState(ctx, &tweenControl{
		fromF: float64(target),
		toF:   float64(target),
	})

	sched := ctx.app.scheduler

	var stepFn func(time.Time) bool
	stepFn = func(now time.Time) bool {
		ctrl.mu.Lock()
		defer ctrl.mu.Unlock()
		if !ctrl.running {
			return false
		}
		t := now.Sub(ctrl.start).Seconds() / duration.Seconds()
		if t >= 1.0 {
			setCurrent(T(ctrl.toF))
			ctrl.fromF = ctrl.toF
			ctrl.running = false
			return false
		}
		v := ctrl.fromF + (ctrl.toF-ctrl.fromF)*easing(t)
		setCurrent(T(v))
		return true
	}

	// Detect a target change and start a new tween. Release ctrl.mu before
	// calling sched.register to avoid lock-order inversion.
	targetF := float64(target)
	var needRegister bool
	ctrl.mu.Lock()
	if targetF != ctrl.toF {
		ctrl.fromF = float64(current)
		ctrl.toF = targetF
		ctrl.start = time.Now()
		ctrl.running = true
		needRegister = true
	}
	ctrl.mu.Unlock()

	if needRegister && sched != nil {
		sched.register(id, stepFn)
	}

	ctx.UseEffect(func() func() {
		return func() {
			if sched != nil {
				sched.unregister(id)
			}
		}
	})

	return current
}

// ── UseTweenColor ────────────────────────────────────────────────────────────

// colorTweenControl holds mutable state shared between a UseTweenColor hook
// and the scheduler.
type colorTweenControl struct {
	mu                  sync.Mutex
	fromR, fromG, fromB float64
	toR, toG, toB       float64
	start               time.Time
	running             bool
}

// rgbComponents returns the red, green, and blue components of c as float64
// values in [0, 255]. Returns (0, 0, 0) for non-RGB named palette colours.
func rgbComponents(c tcell.Color) (r, g, b float64) {
	if !c.IsRGB() {
		return 0, 0, 0
	}
	hex := c.Hex()
	return float64((hex >> 16) & 0xff),
		float64((hex >> 8) & 0xff),
		float64(hex & 0xff)
}

// UseTweenColor smoothly interpolates a tcell.Color toward target whenever
// target changes, animating over duration using the given easing function.
//
// Both the current and target colours should be RGB colours created with
// tcell.NewRGBColor. Named palette colours (e.g. tcell.ColorBlue) are not
// interpolatable and will fall back to showing the target immediately.
//
// Example — animated focus border colour:
//
//	target := tcell.ColorGray
//	if focused {
//	    target = tcell.NewRGBColor(0, 120, 215)
//	}
//	borderColor := tz.UseTweenColor(ctx, target, 120*time.Millisecond, tz.EaseOut)
func UseTweenColor(ctx *RenderContext, target tcell.Color, duration time.Duration, easing EasingFn) tcell.Color {
	baseIdx := ctx.hookIndex
	id := animID(ctx, "tween-color", baseIdx)

	toR, toG, toB := rgbComponents(target)
	current, setCurrent := UseState(ctx, target)
	ctrl, _ := UseState(ctx, &colorTweenControl{
		fromR: toR, fromG: toG, fromB: toB,
		toR: toR, toG: toG, toB: toB,
	})

	sched := ctx.app.scheduler

	var stepFn func(time.Time) bool
	stepFn = func(now time.Time) bool {
		ctrl.mu.Lock()
		defer ctrl.mu.Unlock()
		if !ctrl.running {
			return false
		}
		t := now.Sub(ctrl.start).Seconds() / duration.Seconds()
		if t >= 1.0 {
			setCurrent(tcell.NewRGBColor(int32(ctrl.toR), int32(ctrl.toG), int32(ctrl.toB)))
			ctrl.fromR, ctrl.fromG, ctrl.fromB = ctrl.toR, ctrl.toG, ctrl.toB
			ctrl.running = false
			return false
		}
		p := easing(t)
		lerp := func(a, b float64) int32 { return int32(math.Round(a + (b-a)*p)) }
		setCurrent(tcell.NewRGBColor(
			lerp(ctrl.fromR, ctrl.toR),
			lerp(ctrl.fromG, ctrl.toG),
			lerp(ctrl.fromB, ctrl.toB),
		))
		return true
	}

	// Detect target change. Release ctrl.mu before registering.
	var needRegister bool
	ctrl.mu.Lock()
	if toR != ctrl.toR || toG != ctrl.toG || toB != ctrl.toB {
		curR, curG, curB := rgbComponents(current)
		ctrl.fromR, ctrl.fromG, ctrl.fromB = curR, curG, curB
		ctrl.toR, ctrl.toG, ctrl.toB = toR, toG, toB
		ctrl.start = time.Now()
		ctrl.running = true
		needRegister = true
	}
	ctrl.mu.Unlock()

	if needRegister && sched != nil {
		sched.register(id, stepFn)
	}

	ctx.UseEffect(func() func() {
		return func() {
			if sched != nil {
				sched.unregister(id)
			}
		}
	})

	return current
}
