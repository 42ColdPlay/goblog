package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

var router = mux.NewRouter()

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

//ArticleFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	titles := r.FormValue("title")
	body := r.FormValue("body")
	errors := make(map[string]string)
	//将中文字符转换为传统字符来计算长度utf8.RuneCountnString()
	//验证标题
	if titles == "" {
		errors["titles"] = "标题不能为空"
	} else if utf8.RuneCountInString(titles) < 3 || utf8.RuneCountInString(titles) > 40 {
		errors["titles"] = "标题长度需介于3-40"
	}
	//验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于10个字符"
	}
	//检查是否有错误
	if len(errors) == 0 {
		fmt.Fprint(w, "验证通过<br>")
		fmt.Fprintf(w, "title的长度为：%v<br>", len(titles))
		fmt.Fprintf(w, "title的值为：%v<br>", titles)
		fmt.Fprintf(w, "body的长度为：%v<br>", len(body))
		fmt.Fprintf(w, "body的值为：%v<br>", body)
	} else {
		html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<title>创建文章 —— 我的技术博客</title>
		<style type="text/css">.error {color: red;}</style>
	</head>
	<body>
		<form action="{{ .URL }}" method="post">
			<p><input type="text" name="title" value="{{ .Title }}"></p>
			{{ with .Errors.titles }}
			<p class="error">{{ . }}</p>
			{{ end }}
			<p><textarea name="body" cols="30" rows="10">{{ .Body }}</textarea></p>
			{{ with .Errors.body }}
			<p class="error">{{ . }}</p>
			{{ end }}
			<p><button type="submit">提交</button></p>
		</form>
	</body>
	</html>
		`
		storeURL, _ := router.Get("articles.store").URL()
		data := ArticlesFormData{
			Title:  titles,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		tmpl, err := template.New("create-form").Parse(html)
		if err != nil {
			panic(err)
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			panic(err)
		}
	}
}

//创建博文表单
func articlesCreateHandler(w http.ResponseWriter, r *http.Request) {
	html := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<title>创建文章--我的技术博客</title>
	</head>
	<body>
		<form action="%s?test=data" method="post">
			<p><input type="text" name="title"></p>
			<p><textarea name="body" cols="30" rows="10"></textarea></p>
			<p><button type="submit">提交</button></p>
	</body>
	</html>
	`
	storeURL, _ := router.Get("articles.store").URL()
	fmt.Fprintf(w, html, storeURL)
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

//使用中间件，对进来的请求先做处理，然后再传给Gorilla Mux去解析
func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}
func main() {

	//Name() 方法用来给路由命名
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")
	//使用 {name} 花括号来设置路径参数
	//在有正则匹配的情况下，使用 : 区分。第一部分是名称，第二部分是正则表达式
	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")

	//自定义404页面
	router.NotFoundHandler = http.HandlerFunc(notFundHandler)

	//中间件：强制内容类型为HTML
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
