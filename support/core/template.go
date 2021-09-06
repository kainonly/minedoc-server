package core

const modelTpl = `
// Code generated by bit. DO NOT EDIT.

package model

import (
	"database/sql/driver"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
	"time"
)

type Array []interface{}

func (x *Array) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Array) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

type Object map[string]interface{}

func (x *Object) Scan(input interface{}) error {
	return jsoniter.Unmarshal(input.([]byte), x)
}

func (x Object) Value() (driver.Value, error) {
	return jsoniter.Marshal(x)
}

func True() *bool {
	value := true
	return &value
}

func False() *bool {
	return new(bool)
}

{{range .}}
    type {{title .Key}} struct {
` +
	"ID     uint   `json:\"id\"`\n" +
	"Status     *bool      `gorm:\"default:true\" json:\"status\"`\n" +
	"CreateTime time.Time  `gorm:\"autoCreateTime\" json:\"create_time\"`\n" +
	"UpdateTime time.Time  `gorm:\"autoUpdateTime\" json:\"update_time\"`\n" +
	`{{range .Schema.Columns}}` +
	"{{title .Key}} {{typ .Type}} `{{tag .}}`\n" +
	`{{end}}` +
	`}
{{end}}

func AutoMigrate(tx *gorm.DB, models ...string) {
	mapper := map[string]interface{}{
		{{range .}}"{{ .Key}}": &{{title .Key}}{},{{end}}
	}

	for _, model := range models {
		if mapper[model] != nil {
			tx.AutoMigrate(mapper[model])
		}
	}
}
`
