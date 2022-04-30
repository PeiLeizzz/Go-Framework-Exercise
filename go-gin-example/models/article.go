package models

import "github.com/jinzhu/gorm"

type Article struct {
	Model

	// gorm 约定通过 类名+ID 的方式去查找两个类的映射关系(belongs to)
	TagID int `json:"tag_id" gorm:"index"` // 用于声明这个字段为索引，如果使用了自动迁移功能则会有影响，否则无影响
	Tag   Tag `json:"tag"`

	Title         string `json:"title"`
	Desc          string `json:"desc"`
	Content       string `json:"content"`
	CreatedBy     string `json:"created_by"`
	ModifiedBy    string `json:"modified_by"`
	State         int    `json:"state"`
	CoverImageUrl string `json:"cover_image_url"`
}

func ExistArticleByID(id int) (bool, error) {
	var article Article
	err := db.Select("id").Where("id = ? and deleted_on = ?", id, 0).First(&article).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	return article.ID > 0, nil
}

func GetArticleTotal(maps interface{}) (int, error) {
	var count int
	err := db.Model(&Article{}).Where(maps).Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	return count, nil
}

func GetArticles(pageNum int, pageSize int, maps interface{}) ([]*Article, error) {
	// 查询 articles 时加载相关的 tags
	// 1. SELECT * FROM blog_articles -> tag_id
	// 2. SELECT * FROM blog_tag WHERE id IN ($tag_id)
	var articles []*Article
	var err error

	if pageNum > 0 && pageSize > 0 {
		err = db.Preload("Tag").Where(maps).Offset(pageNum).Limit(pageSize).Find(&articles).Error
	} else {
		err = db.Preload("Tag").Where(maps).Find(&articles).Error
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return articles, nil
}

func GetArticle(id int) (*Article, error) {
	var article Article
	err := db.Where("id = ? and deleted_on = ?", id, 0).First(&article).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// `Related` usually used when you already loaded the User
	// `Association` usually used when you need to do more advanced tasks
	err = db.Model(&article).Related(&article.Tag).Error
	// db.Model(&article).Association("Tag").Find(&article.Tag)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &article, nil
}

func EditArticle(id int, data interface{}) error {
	err := db.Model(&Article{}).Where("id = ? and deleted_on = ?", id, 0).Updates(data).Error

	return err
}

func AddArticle(data map[string]interface{}) error {
	err := db.Create(&Article{
		TagID:         data["tag_id"].(int),
		Title:         data["title"].(string),
		Desc:          data["desc"].(string),
		Content:       data["content"].(string),
		CreatedBy:     data["created_by"].(string),
		State:         data["state"].(int),
		CoverImageUrl: data["cover_image_url"].(string),
	}).Error

	return err
}

func DeleteArticle(id int) error {
	err := db.Where("id = ? and deleted_on = ?", id, 0).Delete(&Article{}).Error

	return err
}

func CleanAllArticles() error {
	err := db.Unscoped().Where("deleted_on != ?", 0).Delete(&Article{}).Error

	return err
}
