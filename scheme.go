package web

import (
	"github.com/tiaotiao/mapstruct"
)

// Tag name for Scheme.
var SchemeTag = "web"

func Scheme(vals map[string]interface{}, dst interface{}) (err error) {
	err = mapstruct.Map2StructTag(vals, dst, SchemeTag)
	if err != nil {
		return NewErrorMsg("invalid argument", err.Error(), StatusBadRequest)
	}
	return nil
}
