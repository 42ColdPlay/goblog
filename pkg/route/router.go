package route

import (
	"github.com/gorilla/mux"
)

//Router路由对象
var Router *mux.Router

//初始化路由
func Initialize() {
	Router = mux.NewRouter()
}

//Name2URL 通过路由名称来获取URL
func Name2URL(routeName string, pairs ...string) string {
	url, err := Router.Get(routeName).URL(pairs...)
	if err != nil {
		return ""
	}
	return url.String()
}
