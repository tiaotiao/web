package web

import (
	"github.com/tiaotiao/mapstruct"
)

var StructTag = "web"

func Scheme(vals map[string]interface{}, dst interface{}) (err error) {
	err = mapstruct.Map2StructTag(vals, dst, StructTag)
	if err != nil {
		return NewErrorMsg("invalid argument", err.Error(), StatusBadRequest)
	}
	return nil
}
