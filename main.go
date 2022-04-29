package main

import (
	"fmt"
	"net/http"
	"strings"
)

/*
func handlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>hello 这里是 goblog</h1>")
	fmt.Fprint(w,"请求路径为："+r.URL.Path)
}
*/
func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.URL.Path == "/" {
		fmt.Fprint(w, "<h1>hello 这里是 goblog!</h1>")
	} else {
		w.WriteHeader(http.StatusNotFound) //添加状态码
		fmt.Fprint(w, "<h1>请求页面未找到！</h1>")
	}
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "此博客是用以记录编程笔记！"+"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}
func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", handlerFunc)
	router.HandleFunc("/about", aboutHandler)
	//文章详情
	router.HandleFunc("/articles/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.SplitN(r.URL.Path, "/", 3)[2]
		fmt.Fprint(w, "文章ID："+id)
	})
	//列表 or 创建
	router.HandleFunc("/wenzhang", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprint(w, "访问文章列表")
		case "POST":
			fmt.Fprint(w, "创建新的文章")
		}
	})
	http.ListenAndServe(":3000", router)
}
