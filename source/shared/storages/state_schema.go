package storages

type MetaRecord struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}

func (MetaRecord) TableName() string { return "meta" }
