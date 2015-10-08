package web

import (
	"fmt"
	"github.com/tiaotiao/mapstruct"
)

// Tag name for Scheme.
var SchemeTagName = "web"

func Scheme(vals map[string]interface{}, dst interface{}) (err error) {
	err = mapstruct.Map2StructTag(vals, dst, SchemeTagName)
	if err != nil {
		return NewErrorMsg("invalid argument", err.Error(), StatusBadRequest)
	}
	return nil
}

func SchemeParam(vals map[string]interface{}, dst interface{}, tag string) (err error) {
	err = mapstruct.Map2Field(vals, dst, tag)
	if err != nil {
		return NewErrorMsg("invalid argument", fmt.Sprintf("'%v' %v", tag, err.Error()), StatusBadRequest)
	}
	return nil
}
