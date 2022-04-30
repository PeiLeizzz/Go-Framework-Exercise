package article_service

import (
	"encoding/json"

	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/PeiLeizzz/go-gin-example/pkg/gredis"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/service/cache_service"
)

type Article struct {
	ID            int
	TagID         int
	Title         string
	Desc          string
	Content       string
	CoverImageUrl string
	State         int
	CreateBy      string
	ModifiedBy    string

	PageNum  int
	PageSize int
}

func (a *Article) Get() (*models.Article, error) {
	var article *models.Article

	cacheArticle := cache_service.Article{ID: a.ID}
	key := cacheArticle.GetArticleKey()
	if gredis.Exists(key) {
		data, err := gredis.Get(key)
		if err != nil {
			logging.Info(err)
		} else {
			// 就算 article 是指针，也要传引用
			json.Unmarshal(data, &article)
			return article, nil
		}
	}

	article, err := models.GetArticle(a.ID)
	if err != nil {
		return nil, err
	}

	gredis.Set(key, article, 3600)
	return article, nil
}

func (a *Article) GetAll() ([]*models.Article, error) {
	var articles []*models.Article

	cacheArticle := cache_service.Article{
		TagID: a.TagID,
		State: a.State,

		PageNum:  a.PageNum,
		PageSize: a.PageSize,
	}
	key := cacheArticle.GetArticlesKey()
	if gredis.Exists(key) {
		data, err := gredis.Get(key)
		if err != nil {
			logging.Info(err)
		} else {
			json.Unmarshal(data, &articles)
			return articles, nil
		}
	}

	articles, err := models.GetArticles(a.PageNum, a.PageSize, a.selectMaps())
	if err != nil {
		return nil, err
	}

	gredis.Set(key, articles, 3600)
	return articles, nil
}

func (a *Article) Add() error {
	article := map[string]interface{}{
		"tag_id":          a.TagID,
		"title":           a.Title,
		"desc":            a.Desc,
		"content":         a.Content,
		"created_by":      a.CreateBy,
		"cover_image_url": a.CoverImageUrl,
		"state":           a.State,
	}
	return models.AddArticle(article)
}

func (a *Article) Edit() error {
	article := map[string]interface{}{
		"tag_id":      a.TagID,
		"title":       a.Title,
		"desc":        a.Desc,
		"content":     a.Content,
		"modified_by": a.ModifiedBy,
	}
	if a.State >= 0 {
		article["state"] = a.State
	}
	if a.CoverImageUrl != "" {
		article["cover_image_url"] = a.CoverImageUrl
	}
	return models.EditArticle(a.ID, article)
}

func (a *Article) Delete() error {
	return models.DeleteArticle(a.ID)
}

func (a *Article) ExistByID() (bool, error) {
	return models.ExistArticleByID(a.ID)
}

func (a *Article) Count() (int, error) {
	return models.GetArticleTotal(a.selectMaps())
}

// 仅用于查询
func (a *Article) selectMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	maps["deleted_on"] = 0
	if a.State >= 0 {
		maps["state"] = a.State
	}
	if a.TagID >= 0 {
		maps["tag_id"] = a.TagID
	}

	return maps
}
