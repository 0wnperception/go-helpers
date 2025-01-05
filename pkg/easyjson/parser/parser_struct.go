package parser

import "github.com/0wnperception/go-helpers/pkg/easyjson"

// easyjson:json
type Model[T easyjson.MarshalerUnmarshaler, V any] struct {
	Data  T      `json:"data"`
	Data2 V      `json:"data2"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}
