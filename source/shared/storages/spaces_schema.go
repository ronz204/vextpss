package storages

import "time"

type SpaceRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (SpaceRecord) TableName() string { return "spaces" }
