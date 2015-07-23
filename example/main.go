package main

import (
	"flag"
)

var addr = flag.String("addr", ":8088", "listen tcp addr")

func main() {
	flag.Parse()

	s := NewServer(*addr)

	err := s.Run()

	if err != nil {
		println("error:", err.Error())
	}
}
