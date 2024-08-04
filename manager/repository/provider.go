package repository

import "gorm.io/gorm"

type ProxyProvider struct {
	gorm.Model
	Name       string
	ProviderId string `gorm:"uniqueIndex;type:varchar(32)"`
}
