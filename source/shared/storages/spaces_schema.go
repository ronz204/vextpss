package storages

import (
	"time"

	"vextpss/source/secrets/core"
)

type SpaceRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SpaceRecord) TableName() string { return "spaces" }

func toSpace(r SpaceRecord) core.Space {
	return core.Space{
		ID:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt,
	}
}
