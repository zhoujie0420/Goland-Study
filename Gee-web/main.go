package main

import (
	"Gee-web/gee"
	"fmt"
	"net/http"
)

func main() {
	//engine := new(Engine)
	////启动web服务，地址+端口
	//log.Fatal(http.ListenAndServe(":9999", engine))

	r := gee.New()
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "url = %q\n", req.URL.Path)
	})

	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "handler[%q] = %q\n", k, v)
		}
	})

	r.Run(":9999")
}
