package tag_service

import (
	"encoding/json"

	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/PeiLeizzz/go-gin-example/pkg/gredis"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/service/cache_service"
)

type Tag struct {
	ID         int
	Name       string
	CreatedBy  string
	ModifiedBy string
	State      int

	PageNum  int
	PageSize int
}

func (t *Tag) GetAll() ([]*models.Tag, error) {
	var tags []*models.Tag

	cacheTag := cache_service.Tag{
		Name:     t.Name,
		State:    t.State,
		PageNum:  t.PageNum,
		PageSize: t.PageSize,
	}
	key := cacheTag.GetTagsKey()
	if gredis.Exists(key) {
		data, err := gredis.Get(key)
		if err != nil {
			logging.Info(err)
		} else {
			json.Unmarshal(data, &tags)
			return tags, nil
		}
	}

	tags, err := models.GetTags(t.PageNum, t.PageSize, t.selectMaps())
	if err != nil {
		return nil, err
	}

	gredis.Set(key, tags, 3600)
	return tags, nil
}

func (t *Tag) Get() (*models.Tag, error) {
	var tag *models.Tag

	cacheTag := cache_service.Tag{ID: t.ID}
	key := cacheTag.GetTagKey()
	if gredis.Exists(key) {
		data, err := gredis.Get(key)
		if err != nil {
			logging.Info(err)
		} else {
			json.Unmarshal(data, &tag)
			return tag, nil
		}
	}

	tag, err := models.GetTag(t.ID)
	if err != nil {
		return nil, err
	}

	gredis.Set(key, tag, 3600)
	return tag, nil
}

func (t *Tag) Add() error {
	tag := map[string]interface{}{
		"name":       t.Name,
		"state":      t.State,
		"created_by": t.CreatedBy,
	}
	return models.AddTag(tag)
}

func (t *Tag) Edit() error {
	tag := map[string]interface{}{
		"name":        t.Name,
		"modified_by": t.ModifiedBy,
	}
	if t.State >= 0 {
		tag["state"] = t.State
	}
	return models.EditTag(t.ID, tag)
}

func (t *Tag) Delete() error {
	return models.DeleteTag(t.ID)
}

func (t *Tag) Count() (int, error) {
	return models.GetTagTotal(t.selectMaps())
}

func (t *Tag) ExistByID() (bool, error) {
	return models.ExistTagByID(t.ID)
}

func (t *Tag) ExistByName() (bool, error) {
	return models.ExistTagByName(t.Name)
}

func (t *Tag) selectMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	maps["deleted_on"] = 0

	if t.Name != "" {
		maps["name"] = t.Name
	}
	if t.State >= 0 {
		maps["state"] = t.State
	}

	return maps
}
