package repository

import (
	"gorm.io/gorm"
)

type Service struct {
	Host   string
	Name   string
	Params map[string]string `gorm:"serializer:json"`
}

type ProxyApi struct {
	gorm.Model
	Service        Service `gorm:"embedded;embeddedPrefix:service_"`
	UpdateInterval float64
	Name           string
	ProviderId     string `gorm:"type:varchar(32);index"`
	ApiId          string `gorm:"uniqueIndex;type:varchar(32)"`
}
