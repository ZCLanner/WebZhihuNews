package main

import (
	"encoding/json"
	"fmt"
	"github.com/lunny/xorm"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"net/http"
)

type NewsError struct {
	description string
}

func (err *NewsError) Error() string {
	return err.description
}

type NewsIndex struct {
	Id        int
	Thumbnail string
	GaPrefix  string
	Title     string
}

type LatestIndices struct {
	Date string
	News []NewsIndex
}

type News struct {
	Id      int
	Content string
}

func getLatestIndices() (*LatestIndices, error) {
	resp, err := http.Get("http://news.at.zhihu.com/api/1.1/news/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var indices LatestIndices
	json.Unmarshal(body, &indices)

	return &indices, nil
}

func getArticle(id int) (string, error) {
	baseUrl := "http://daily.zhihu.com/api/1.1/news/"
	url := fmt.Sprintf("%s%d", []byte(baseUrl), id)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func Crawl(engine *xorm.Engine) error {
	latestIndices, err := getLatestIndices()
	if err != nil {
		return err
	}

	_, err = engine.Insert(&latestIndices.News)
	if err != nil {
		return err
	}

	news := make([]News, 0)
	for _, newsIndex := range latestIndices.News {
		content, err := getArticle(newsIndex.Id)
		if err == nil {
			news = append(news, News{newsIndex.Id, content})
		}
	}
	_, err = engine.Insert(&news)
	if err != nil {
		return err
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
