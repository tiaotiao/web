package web

import (
	//"fmt"
	"reflect"
	"testing"
)

type Item struct {
	Name string `web:"slice"`
}

func TestScheme(t *testing.T) {

	var obj = struct {
		ReqID    uint32 `web:"id,required"`
		Name     string `web:"name"`
		Default  uint   `web:",20"`
		Zero     int64
		Ignore   string `web:"-"`
		NotFound string `web:"notfound"`

		Bool    bool
		Int     int
		Int8    int8 `name:"uid" desc:"spec userid"`
		Int16   int16
		Int32   int32
		Int64   int64
		Uint8   uint8
		Uint16  uint16
		Uint32  uint32
		Uint64  uint64
		Float   float32
		Float64 float64

		ItemObj      Item    `web:"item"`
		ItemPtr      *Item   `web:"itemptr"`
		ItemSlice    []Item  `web:"itemslice"`
		ItemPtrSlice []*Item `web:"itemptrslice"`

		StrSlice   []string
		IntSlice   []int64
		OneSlice   []int
		EmptySlice []string

		// ignored
		anonymous string `web:",required"`

		// not support, just ignored
		Map map[int]int
		//Slice   []int
		Struct  struct{ A int }
		Pointer *int
	}{}
	obj.NotFound = "not changed"

	chk := obj
	chk.ReqID = 123465
	chk.Name = "tom"
	chk.Default = 20
	chk.Zero = 0
	chk.Ignore = ""
	chk.NotFound = "not changed"
	chk.Int = 20
	chk.Int8 = 127
	chk.Int16 = -32768
	chk.Int32 = -20
	chk.Int64 = int64(9000000000000000001)
	chk.Uint8 = 255
	chk.Uint16 = 65535
	chk.Uint32 = 4294967295
	chk.Uint64 = uint64(9300000000000000001)
	chk.Float = 10.3
	chk.Float64 = 1000000.00000001
	chk.Bool = true

	chk.ItemObj = Item{Name: "itemname"}
	chk.ItemPtr = &Item{Name: "itemname"}
	chk.ItemSlice = []Item{{Name: "item1"}, {Name: "item2"}}
	chk.ItemPtrSlice = []*Item{{Name: "item1"}}

	chk.StrSlice = []string{"abc", "123", "efg"}
	chk.IntSlice = []int64{10, 20, 30}
	chk.OneSlice = []int{50}
	chk.EmptySlice = nil

	var vals = make(map[string]interface{})
	vals["id"] = "123465"
	vals["name"] = "tom"
	//vals["default"]
	//vals["zero"]
	vals["ignore"] = "x$%^&*(HJ"
	//vals["notfound"]
	vals["int"] = "20"
	vals["int8"] = 127
	vals["int16"] = -32768
	vals["int32"] = -20
	vals["int64"] = int64(9000000000000000001)
	vals["uint8"] = 255
	vals["uint16"] = 65535
	vals["uint32"] = uint32(4294967295)
	vals["uint8"] = 255
	vals["uint16"] = 65535
	vals["uint32"] = uint32(4294967295)
	vals["uint64"] = "9300000000000000001"
	vals["float"] = "10.3"
	vals["float64"] = 1000000.00000001
	vals["bool"] = "true"

	vals["item"] = "{\"name\":\"itemname\"}"
	vals["itemptr"] = "{\"name\":\"itemname\"}"
	vals["itemslice"] = `[{"name":"item1"}, {"name":"item2"}]`
	vals["itemptrslice"] = `[{"name":"item1"}]`

	vals["strslice"] = "abc,123,efg"
	vals["intslice"] = "10,20,30"
	vals["oneslice"] = "50"
	vals["emptyslice"] = ""

	// scheme convert
	err := Scheme(vals, &obj)

	if err != nil {
		t.Fatal("ERROR: ", err.Error())
	}

	if !reflect.DeepEqual(obj, chk) {
		t.Fatalf("\nobj =%v\nwant=%v\n", obj, chk)
	}

	// data type mismatch
	obj2 := struct {
		Uid int `web:"fid"`
	}{}

	val := map[string]interface{}{"fid": "abc"}
	err = Scheme(val, &obj2)
	if err == nil {
		t.Fatalf("map type 'str' struct type 'int' but not err msg")
	}

	// required data lack
	obj3 := struct {
		Name    string `web:"name,required"`
		Content string
	}{}
	val = map[string]interface{}{"content": "t"}
	err = Scheme(val, &obj3)
	if err == nil {
		t.Fatal("struct need required element lack but not err msg", val)
	}

	// out range of int
	obj4 := struct {
		Cint8 int8
	}{}
	val = map[string]interface{}{"cint8": "210"}
	err = Scheme(val, &obj4)
	if err != nil {
		t.Fatal(err)
	}

	// map element more than struct
	obj5 := struct {
		F1 string
	}{}
	val = map[string]interface{}{"f1": "v1", "f2": "v2"}
	err = Scheme(val, &obj5)
	if err != nil {
		t.Fatal(err)
	}
}
