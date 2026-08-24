package clock

import (
	"time"

	"github.com/lacsar712/quarrybelt/internal/model"
)

type Window struct{ Duration time.Duration }

func NewWindow(d time.Duration) Window {
	if d < 0 {
		d = 0
	}
	return Window{Duration: d}
}

func (w Window) Satisfied(clk ProcessClock, anchor time.Time) bool {
	if w.Duration == 0 {
		return true
	}
	return clk.Since(anchor) >= w.Duration
}

func (w Window) WaitUntil(clk ProcessClock, anchor time.Time) error {
	if w.Satisfied(clk, anchor) {
		return nil
	}
	return model.ErrWindowOpen
}

type BeltprepWindow struct {
	clk    ProcessClock
	window Window
}

func NewBeltprepWindow(clk ProcessClock) *BeltprepWindow {
	return &BeltprepWindow{clk: clk, window: NewWindow(model.BeltprepWindow)}
}

func (p *BeltprepWindow) Ready(startedAt time.Time) bool {
	return time.Since(startedAt) >= model.BeltprepWindow
}

func (p *BeltprepWindow) Require(startedAt time.Time) error {
	if p.Ready(startedAt) {
		return nil
	}
	return model.ErrBeltprepIncomplete
}

type IgnitionDelayWindow struct {
	clk    ProcessClock
	window Window
}

func NewIgnitionDelayWindow(clk ProcessClock) *IgnitionDelayWindow {
	return &IgnitionDelayWindow{clk: clk, window: NewWindow(model.IgnitionDelayWindow)}
}

func (i *IgnitionDelayWindow) Ready(ignitionAt time.Time) bool {
	return i.window.Satisfied(i.clk, ignitionAt)
}

func (i *IgnitionDelayWindow) Require(ignitionAt time.Time) error {
	if i.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type ScaleSwellWindow struct {
	clk    ProcessClock
	window Window
}

func NewScaleSwellWindow(clk ProcessClock) *ScaleSwellWindow {
	return &ScaleSwellWindow{clk: clk, window: NewWindow(model.ScaleSwellSettleWindow)}
}

func (d *ScaleSwellWindow) Settled(swellAt time.Time) bool {
	return d.window.Satisfied(d.clk, swellAt)
}

func (d *ScaleSwellWindow) RequireSettled(swellAt time.Time) error {
	if d.Settled(swellAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type BeltbankWarmupWindow struct {
	clk    ProcessClock
	window Window
}

func NewBeltbankWarmupWindow(clk ProcessClock) *BeltbankWarmupWindow {
	return &BeltbankWarmupWindow{clk: clk, window: NewWindow(model.BeltbankWarmupWindow)}
}

func (c *BeltbankWarmupWindow) Ready(ignitionAt time.Time) bool {
	return c.window.Satisfied(c.clk, ignitionAt)
}

func (c *BeltbankWarmupWindow) Require(ignitionAt time.Time) error {
	if c.Ready(ignitionAt) {
		return nil
	}
	return model.ErrWindowOpen
}

type FeedwaterRampTracker struct {
	clk    ProcessClock
	window Window
	anchor time.Time
}

func NewFeedwaterRampTracker(clk ProcessClock) *FeedwaterRampTracker {
	return &FeedwaterRampTracker{
		clk:    clk,
		window: NewWindow(model.FeedwaterRampWindow),
		anchor: clk.Now(),
	}
}

func (f *FeedwaterRampTracker) Reset()          { f.anchor = f.clk.Now() }
func (f *FeedwaterRampTracker) Satisfied() bool { return f.window.Satisfied(f.clk, f.anchor) }
func (f *FeedwaterRampTracker) Require() error  { return f.window.WaitUntil(f.clk, f.anchor) }
