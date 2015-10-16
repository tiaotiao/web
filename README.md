web [![GoDoc](https://godoc.org/github.com/gocraft/web?status.png)](https://godoc.org/github.com/tiaotiao/web)
=======================
tiaotiao/web is a web server framework of golang.

Getting Started
-------------------------
Install:

	# go get github.com/tiaotiao/web
  
A simple HelloWorld example:
```
package main

import (
	github.com/tiaotiao/web
)

func Hello(c *web.Context) interface{} {
	return "hello world" // return a string will be write directly with out processing.
}

func main() {
	w := web.NewWeb()
	w.Handle("GET", "/api/hello", Hello)
	w.ListenAndServe("tcp", ":8080")  // block forever until closed
}
```

Features
------------------------------------------

	* Router
	* Context
	* Middleware
	* ParseParams
	* Scheme
	* Response

