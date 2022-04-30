package models

import "github.com/jinzhu/gorm"

type Tag struct {
	Model

	Name       string `json:"name"`
	CreatedBy  string `json:"created_by"`
	ModifiedBy string `json:"modified_by"`
	State      int    `json:"state"`
}

func GetTag(id int) (*Tag, error) {
	var tag Tag
	err := db.Where("id = ? and deleted_on = ?", id, 0).First(&tag).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return &tag, nil
}

func GetTags(pageNum int, pageSize int, maps interface{}) ([]*Tag, error) {
	var tags []*Tag
	var err error

	if pageSize > 0 && pageNum > 0 {
		err = db.Where(maps).Offset(pageNum).Limit(pageSize).Find(&tags).Error
	} else {
		err = db.Where(maps).Find(&tags).Error
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return tags, nil
}

func GetTagTotal(maps interface{}) (int, error) {
	var count int
	err := db.Model(&Tag{}).Where(maps).Count(&count).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return 0, err
	}

	return count, nil
}

func AddTag(data map[string]interface{}) error {
	err := db.Create(&Tag{
		Name:      data["name"].(string),
		State:     data["state"].(int),
		CreatedBy: data["created_by"].(string),
	}).Error

	return err
}

func ExistTagByName(name string) (bool, error) {
	var tag Tag
	err := db.Select("id").Where("name = ? and deleted_on = ?", name, 0).First(&tag).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	return tag.ID > 0, nil
}

func ExistTagByID(id int) (bool, error) {
	var tag Tag
	err := db.Select("id").Where("id = ? and deleted_on = ?", id, 0).First(&tag).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	return tag.ID > 0, nil
}

func DeleteTag(id int) error {
	err := db.Where("id = ? and deleted_on = ?", id, 0).Delete(&Tag{}).Error

	return err
}

func EditTag(id int, data interface{}) error {
	err := db.Model(&Tag{}).Where("id = ? and deleted_on = ?", id, 0).Updates(data).Error

	return err
}

func CleanAllTags() error {
	// 硬删除要使用 Unscoped() 这是 Gorm 的约定
	// 因为在 model.go 的 Delete Callback 中，对 scope.Search.Unscoped 进行了检查
	err := db.Unscoped().Where("deleted_on != ?", 0).Delete(&Tag{}).Error

	return err
}
