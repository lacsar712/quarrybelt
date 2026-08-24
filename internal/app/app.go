package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/quarrybelt/internal/yard"
	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/beltbank"
	"github.com/lacsar712/quarrybelt/internal/config"
	"github.com/lacsar712/quarrybelt/internal/scale"
	"github.com/lacsar712/quarrybelt/internal/fsm"
	"github.com/lacsar712/quarrybelt/internal/interlock"
	"github.com/lacsar712/quarrybelt/internal/model"
	"github.com/lacsar712/quarrybelt/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.YardFSM
	yard        *yard.Controller
	beltbank    *beltbank.Coordinator
	scale          *scale.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	beltprepWindow   *clock.BeltprepWindow
	warmupWindow  *clock.BeltbankWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	beltLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewYardFSM(cfg.UnitID),
		yard:       yard.NewController(clk),
		beltbank:   beltbank.NewCoordinator(clk),
		scale:         scale.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		beltprepWindow:  clock.NewBeltprepWindow(clk),
		warmupWindow: clock.NewBeltbankWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		beltLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.YardFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetScale(a.scale.Level().WithinLimits(snap.Scale.LevelPercent))
	a.permissives.SetPressure(a.yard.Pressure().WithinTripLimits(snap.Yard.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetBeltbank(a.beltbank.Burner().DriveStable(snap.Beltbank))
	a.permissives.SetBelt(snap.Beltbank.BeltFlowTPH > 0 || snap.State == model.StateBeltprep)
	a.permissives.SetIgnition(snap.Beltbank.BurnerPhase == model.BurnerStable || snap.Beltbank.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetBeltPermissive(a.permissives.BeltOK())
	a.fsm.SetBeltprepComplete(a.beltprepWindow.Ready(snap.Beltbank.BeltprepStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
