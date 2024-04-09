package main

import (
	"fmt"
	"net/http"
)

func sayHelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	print(r.Form)
	print("path", r.URL.Path)
	print("scheme", r.URL.Scheme)
	print(r.Form["url_long"])
	for k, v := range r.Form {
		print("key:", k)
		print("val:", v)
	}
	fmt.Fprintf(w, "Hello astaxie!")
}

func main() {
	http.HandleFunc("/", sayHelloName)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("http server start failed, err:", err)
	}
}
