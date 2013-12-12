package main

import (
	"fmt"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	"go/build"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
)

var baseDirectory string
var engine *xorm.Engine

func main() {
	var err error
	baseDirectory, err = findRoot()
	checkErr(err)

	fmt.Println("base directory =", baseDirectory)
	dbPath := fmt.Sprintf("%s/news.db", []byte(baseDirectory))
	engine, err = xorm.NewEngine(xorm.SQLITE, dbPath)
	if err != nil {
		return
	}
	defer engine.Close()
	engine.ShowSQL = true

	go Crawl(engine)

	router := new(RegexpHandler)

	re, err := regexp.Compile("/css/*")
	checkErr(err)
	dir := fmt.Sprintf("%s/static/", []byte(baseDirectory))
	router.Handler(re, http.FileServer(http.Dir(dir)))

	re, err = regexp.Compile("/js/*")
	checkErr(err)
	router.Handler(re, http.FileServer(http.Dir(dir)))

	re, err = regexp.Compile("/article")
	checkErr(err)
	router.HandleFunc(re, ViewArticle)

	re, err = regexp.Compile("/*")
	checkErr(err)
	router.HandleFunc(re, ListArticles)

	err = http.ListenAndServe(":8080", router)
	checkErr(err)

}

func ListArticles(w http.ResponseWriter, r *http.Request) {
	templateFilePath := fmt.Sprintf("%s/views/templates/listArticle.html", []byte(baseDirectory))
	t, err := template.ParseFiles(templateFilePath)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	var articleIndices []NewsIndex
	err = engine.OrderBy("Id DESC").Find(&articleIndices)
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.Execute(w, articleIndices)
}

func ViewArticle(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	stringIds := r.Form["articleid"]
	if len(stringIds) < 1 {
		fmt.Fprintf(w, "Error")
		return
	}

	id, err := strconv.Atoi(stringIds[0])
	if err != nil {
		fmt.Fprintf(w, "Error")
		return
	}

	var news News
	_, err = engine.Where("id=?", id).Get(&news)
	if err == nil {
		fmt.Fprintf(w, news.Content)
	} else {
		fmt.Fprintf(w, "Error")
	}
}

type URLRouter struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*URLRouter
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &URLRouter{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &URLRouter{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("base directory =", baseDirectory)
	for _, router := range h.routes {
		if router.pattern.MatchString(r.URL.Path) {
			router.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

func findRoot() (string, error) {
	ctx := build.Default
	p, err := ctx.Import("github.com/ZCLanner/WebZhihuNews/", "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}
