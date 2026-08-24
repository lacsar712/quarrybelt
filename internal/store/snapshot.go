package store

import "github.com/lacsar712/quarrybelt/internal/model"

type ScaleSnapshotView struct {
	UnitID   string
	Scale     model.ScaleReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneScaleSnapshot(s model.PlantSnapshot) ScaleSnapshotView {
	out := ScaleSnapshotView{
		UnitID:   s.UnitID,
		Scale:     s.Scale,
		Revision: s.Revision,
	}
	// Copy the backing array, not just the slice header: a three-index
	// expression (s.Alarms[:len:len]) would alias the source's array, so
	// in-place edits to the returned view would leak back into the live
	// snapshot the view was derived from. append(nil, ...) performs a real
	// element copy while still clamping capacity to len.
	out.Alarms = append([]model.AlarmEvent(nil), s.Alarms...)
	return out
}
