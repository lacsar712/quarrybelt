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
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
