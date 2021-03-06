package main

import (
	"database/sql"
	"fmt"
	"goblog/pkg/logger"
	"goblog/pkg/route"
	"goblog/pkg/types"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var router *mux.Router
var db *sql.DB

func initDB() {
	var err error
	config := mysql.Config{
		User:                 "root",
		Passwd:               "gong97826",
		Addr:                 "127.0.0.1:3306",
		Net:                  "tcp",
		DBName:               "goblog",
		AllowNativePasswords: true,
	}
	//准备数据库连接池
	db, err = sql.Open("mysql", config.FormatDSN())
	logger.LogError(err)
	//设置最大连接数
	db.SetMaxOpenConns(25)
	//设置最大空闲连接数
	db.SetMaxIdleConns(25)
	//设置每个连接的过期时间
	db.SetConnMaxLifetime(5 * time.Minute)

	//尝试连接，失败报错
	err = db.Ping()
	logger.LogError(err)

}

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

type Article struct {
	Title, Body string
	ID          int64
}

func articlesShowHandler(w http.ResponseWriter, r *http.Request) {
	//mux.Vars(r) 会将 URL 路径参数解析为键值对应的 Map，使用以下方法即可读取
	//1.获取URL参数
	// vars := mux.Vars(r)
	// id := vars["id"]
	// fmt.Fprint(w, "文章ID:"+id)
	id := route.GetRouteVariable("id", r)
	fmt.Fprint(w, "文章ID:"+id)
	//2.读取对应的文章数据
	//article := Article{}
	//query := "SELECT * FROM articles WHERE id = ?"
	//QueryRow()封装了Prepare方法的调用
	//err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	//上述代码相当于
	/*stmt,err :=db.Prepare(query)
	logger.LogError(err)
	defer stmt.Close()
	err = stmt.QueryRow(id).Scan(&article.ID,&article.Title,&article.Body)
	*/
	article, err := getArticleById(id)

	//3.如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404文章未找到")
		} else {
			//3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500服务器内部错误")
		}
	} else {
		//4.读取成功
		//fmt.Fprint(w, "读取成功，文章标题--"+article.Title)
		//4.读取成功，显示文章
		tmpl, err := template.New("show.gohtml").
			Funcs(template.FuncMap{
				"RouteName2URL": route.Name2URL,
				"Int64ToString": types.Int64ToString,
			}).
			ParseFiles("resources/views/articles/show.gohtml")
		logger.LogError(err)
		err = tmpl.Execute(w, article)
		logger.LogError(err)
	}
}

func articlesIndexHandler(w http.ResponseWriter, r *http.Request) {
	//1.执行查询语句，返回一个结果集
	rows, err := db.Query("SELECT * from articles")
	logger.LogError(err)
	defer rows.Close()

	var articles []Article
	//2.循环读取结果
	for rows.Next() {
		var article Article
		//2.1 扫描每一行的结果并赋值到一个article对象中
		err := rows.Scan(&article.ID, &article.Title, &article.Body)
		logger.LogError(err)
		//2.2 将article追加到articles这个数组切片中
		articles = append(articles, article)
	}
	//2.3.检查遍历时是否发生错误
	err = rows.Err()
	logger.LogError(err)

	//3.加载模板
	tmpl, err := template.ParseFiles("resources/views/articles/index.gohtml")
	logger.LogError(err)

	//4.渲染模板，将所有的文章数据传输进去
	err = tmpl.Execute(w, articles)
	logger.LogError(err)
}

//Link方法用来生成文章链接
func (a Article) Link() string {
	showURL, err := router.Get("articles.show").URL("id", strconv.FormatInt(a.ID, 10))
	if err != nil {
		logger.LogError(err)
		return ""
	}
	return showURL.String()
}

//ArticleFormData 创建博文表单数据
type ArticlesFormData struct {
	Title, Body string
	URL         *url.URL
	Errors      map[string]string
}

func saveArticleToDB(title string, body string) (int64, error) {
	//变量初始化
	var (
		id   int64
		err  error
		rs   sql.Result
		stmt *sql.Stmt
	)
	//1.获取一个prepare声明语句，Prepare语句可以有效防范SQL注入攻击（有效且必备）
	stmt, err = db.Prepare("INSERT INTO articles (title,body) VALUES(?,?)")
	//例行错误检查
	if err != nil {
		return 0, err
	}
	//2.在此函数运行结束之后关闭该语句，防止占用SQL连接
	defer stmt.Close()

	//3.执行请求，传参进入绑定的内容
	rs, err = stmt.Exec(title, body)
	if err != nil {
		return 0, err
	}
	//4.如果传入成功的话，会返回自增ID
	if id, err = rs.LastInsertId(); id > 0 {
		return id, nil
	}
	return 0, err
}
func articlesStoreHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	body := r.FormValue("body")
	errors := make(map[string]string)
	//将中文字符转换为传统字符来计算长度utf8.RuneCountnString()
	//验证标题
	/*if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于3-40"
	}
	//验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于或等于10个字符"
	}*/
	errors = validateArticleFormData(title, body)
	//检查是否有错误
	if len(errors) == 0 {
		lastInsertID, err := saveArticleToDB(title, body)
		if lastInsertID > 0 {
			fmt.Fprint(w, "插入成功,ID为"+strconv.FormatInt(lastInsertID, 10))
		} else {
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500服务器内部错误")
		}
	} else {
		storeURL, _ := router.Get("articles.store").URL()
		data := ArticlesFormData{
			Title:  title,
			Body:   body,
			URL:    storeURL,
			Errors: errors,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
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
	storeURL, _ := router.Get("articles.store").URL()
	data := ArticlesFormData{
		Title:  "",
		Body:   "",
		URL:    storeURL,
		Errors: nil,
	}
	tmpl, err := template.ParseFiles("resources/views/articles/create.gohtml")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}

func getArticleById(id string) (Article, error) {
	article := Article{}
	query := "SELECT * FROM articles where id = ?"
	err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	return article, err
}
func articlesEditHandler(w http.ResponseWriter, r *http.Request) {
	//1.获取URL参数
	//vars := mux.Vars(r)
	//id := vars["id"]
	id := route.GetRouteVariable("id", r)
	//2.读取对应的文章数据
	// article := Article{}
	// query := "SELECT * FROM articles WHERE id = ?"
	// err := db.QueryRow(query, id).Scan(&article.ID, &article.Title, &article.Body)
	article, err := getArticleById(id)

	//3.如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			//3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		//4.读取成功，显示表单
		updateURL, _ := router.Get("articles.update").URL("id", id)
		data := ArticlesFormData{
			Title:  article.Title,
			Body:   article.Body,
			URL:    updateURL,
			Errors: nil,
		}
		tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
		logger.LogError(err)
		err = tmpl.Execute(w, data)
		logger.LogError(err)
	}
}

//封装表单验证
func validateArticleFormData(title string, body string) map[string]string {
	errors := make(map[string]string)
	//验证标题
	if title == "" {
		errors["title"] = "标题不能为空"
	} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
		errors["title"] = "标题长度需介于3-40"
	}

	//验证内容
	if body == "" {
		errors["body"] = "内容不能为空"
	} else if utf8.RuneCountInString(body) < 10 {
		errors["body"] = "内容长度需大于等于10个字节"
	}
	return errors
}
func articlesUpdateHandler(w http.ResponseWriter, r *http.Request) {
	//1.获取URL参数
	id := route.GetRouteVariable("id", r)
	//2.读取对应的文章数据
	_, err := getArticleById(id)
	//3.如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			//3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		//4.未出现错误
		title := r.PostFormValue("title")
		body := r.PostFormValue("body")
		errors := make(map[string]string)
		// 验证标题
		/*if title == "" {
			errors["title"] = "标题不能为空"
		} else if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 40 {
			errors["title"] = "标题长度需介于3-40"
		}

		//验证内容
		if body == "" {
			errors["body"] = "内容不能为空"
		} else if utf8.RuneCountInString(body) < 10 {
			errors["body"] = "内容长度需大于等于10个字节"
		}*/
		errors = validateArticleFormData(title, body)
		if len(errors) == 0 {
			//4.2表单验证通过，更新数据
			query := "UPDATE articles SET title=?,body=? WHERE id = ?"
			rs, err := db.Exec(query, title, body, id)
			if err != nil {
				logger.LogError(err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "500 服务器内部错误")
			}
			//更新成功，跳转到文章详情页
			if n, _ := rs.RowsAffected(); n > 0 {
				showURL, _ := router.Get("articles.show").URL("id", id)
				http.Redirect(w, r, showURL.String(), http.StatusFound)
			} else {
				fmt.Fprint(w, "您未做任何修改！")
			}
		} else {
			//4.3 表单验证不通过，显示理由
			updateURL, _ := router.Get("articles.update").URL("id", id)
			date := ArticlesFormData{
				Title:  title,
				Body:   body,
				URL:    updateURL,
				Errors: errors,
			}
			tmpl, err := template.ParseFiles("resources/views/articles/edit.gohtml")
			logger.LogError(err)
			err = tmpl.Execute(w, date)
			logger.LogError(err)
		}
	}
}

func articlesDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//1.获取URL参数
	id := route.GetRouteVariable("id", r)

	//2.读取对应的文章数据
	article, err := getArticleById(id)
	//3.如果出现错误
	if err != nil {
		if err == sql.ErrNoRows {
			//3.1 数据未找到
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "404 文章未找到")
		} else {
			//3.2 数据库错误
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "500 服务器内部错误")
		}
	} else {
		//4.未出现错误，可以进行删除操作
		rowsAffected, err := article.Delete()

		//4.1 发生错误
		if err != nil {
			//SQL报错
			logger.LogError(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "500 服务器内部错误")
		} else {
			//4.2 未发生错误
			if rowsAffected > 0 {
				//重定向到文章列表页
				indexURL, _ := router.Get("articles.index").URL()
				http.Redirect(w, r, indexURL.String(), http.StatusFound)
			} else {
				// edge case
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "404 文章未找到")
			}
		}
	}
}
func (a Article) Delete() (rowsAffected int64, err error) {
	rs, err := db.Exec("DELETE FROM articles WHERE id = " + strconv.FormatInt(a.ID, 10))
	if err != nil {
		return 0, err
	}
	//删除成功，跳转到文章详情页
	if n, _ := rs.RowsAffected(); n > 0 {
		return n, nil
	}
	return 0, nil
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
func createTables() {
	createArticlesSQL := `
	create table if not exists articles(
		id bigint(20) PRIMARY KEY AUTO_INCREMENT NOT NULL,
		title varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
		body longtext COLLATE utf8mb4_unicode_ci
	);`
	_, err := db.Exec(createArticlesSQL)
	logger.LogError(err)
}
func main() {
	initDB()

	createTables()
	route.Initialize()
	router = route.Router
	//Name() 方法用来给路由命名
	router.HandleFunc("/", homeHandler).Methods("GET").Name("home")
	router.HandleFunc("/about", aboutHandler).Methods("GET").Name("about")
	//使用 {name} 花括号来设置路径参数
	//在有正则匹配的情况下，使用 : 区分。第一部分是名称，第二部分是正则表达式
	router.HandleFunc("/articles/{id:[0-9]+}", articlesShowHandler).Methods("GET").Name("articles.show")
	router.HandleFunc("/articles", articlesIndexHandler).Methods("GET").Name("articles.index")
	router.HandleFunc("/articles", articlesStoreHandler).Methods("POST").Name("articles.store")
	router.HandleFunc("/articles/create", articlesCreateHandler).Methods("GET").Name("articles.create")
	router.HandleFunc("/articles/{id:[0-9]+}/edit", articlesEditHandler).Methods("GET").Name("articles.edit")
	router.HandleFunc("/articles/{id:[0-9]+}", articlesUpdateHandler).Methods("POST").Name("articles.update")
	router.HandleFunc("/articles/{id:[0-9]+}/delete", articlesDeleteHandler).Methods("POST").Name("articles.delete")

	//自定义404页面
	router.NotFoundHandler = http.HandlerFunc(notFundHandler)

	//中间件：强制内容类型为HTML
	router.Use(forceHTMLMiddleware)

	http.ListenAndServe(":3000", removeTrailingSlash(router))
}
