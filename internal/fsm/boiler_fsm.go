package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/quarrybelt/internal/model"
)

type YardFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	beltPermissive bool
	beltprepComplete  bool
	hooks          *HookChain
}

func NewYardFSM(unitID string) *YardFSM {
	_ = unitID
	return &YardFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *YardFSM) Hooks() *HookChain { return f.hooks }

func (f *YardFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *YardFSM) SetBeltPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.beltPermissive = ok
}

func (f *YardFSM) SetBeltprepComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.beltprepComplete = ok
}

func (f *YardFSM) BeltPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.beltPermissive
}

func (f *YardFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		if f.hooks != nil {
			_ = f.hooks.RunAfter(ctx, f.state, f.state, event)
		}
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.beltPermissive {
		return f.state, fmt.Errorf("%w", model.ErrBeltPermissive)
	}
	if event == EvBeltprepComplete && !f.beltprepComplete {
		return f.state, fmt.Errorf("%w", model.ErrBeltprepIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *YardFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
