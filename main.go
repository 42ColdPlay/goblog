package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

/*
func handlerFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>hello 这里是 goblog</h1>")
	fmt.Fprint(w,"请求路径为："+r.URL.Path)
}

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
	//router := http.NewServeMux()
	//使用gorilla/mux，功能强大，但是性能有所不及官方HttpRouter
	router := mux.NewRouter()
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
*/
//http.ServeMux的长度优先匹配适用于静态内容
//gorilla/mux 的精准匹配适合动态网站
func homeHandler(w http.ResponseWriter, r *http.Request) {
	//已使用中间件
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "hello 欢迎来到goblog!</h1>")
}
func aboutHandler(w http.ResponseWriter, r *http.Request) {
	//已使用中间件
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "此博客是用以记录编程笔记！"+"<a href=\"mailto:summer@example.com\">summer@example.com</a>")
}
func notFundHandler(w http.ResponseWriter, r *http.Request) {
	//已使用中间件
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>请求页面未找到 :(</h1><p>如有疑惑，请联系我们。</p>")
}
func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	//mux.Vars(r) 会将 URL 路径参数解析为键值对应的 Map，使用以下方法即可读取
	vars := mux.Vars(r)
	id := vars["id"]
	fmt.Fprint(w, "文章ID："+id)
}
func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "访问文章列表")
}
func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "创建新的文章")
}

//中间件
func forceHTMLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//设置标头
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//2.继续处理请求
		next.ServeHTTP(w, r)
	})
}
func main() {
	router := mux.NewRouter()
	//Name() 方法用来给路由命名
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")
	//使用 {name} 花括号来设置路径参数
	//在有正则匹配的情况下，使用 : 区分。第一部分是名称，第二部分是正则表达式
	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")

	//自定义404页面
	router.NotFoundHandler = http.HandlerFunc(notFundHandler)

	//中间件：强制内容类型为HTML
	router.Use(forceHTMLMiddleware)

	//通过命名路由获取URL示例
	//传参是路由的名称，接下来我们就可以靠这个名称来获取到 URI
	homeURL, _ := router.Get("home").URL()
	fmt.Println("homeURL:", homeURL)
	articleURL, _ := router.Get("articles.show").URL("id", "3")
	fmt.Println("articleURL:", articleURL)
	http.ListenAndServe(":3000", router)
}
