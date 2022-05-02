package tag_service

import (
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/PeiLeizzz/go-gin-example/pkg/export"
	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/gredis"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/service/cache_service"
	"github.com/tealeg/xlsx"
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

	// 注意：name 为空 或者 state < 0 的条件会自动被省略掉
	tags, err := models.GetTags(t.PageNum, t.PageSize, t.selectMaps())
	if err != nil {
		return nil, err
	}

	gredis.Set(key, tags, 3600)
	return tags, nil
}

func (t *Tag) Get() (*models.Tag, error) {
	tag := &models.Tag{}

	cacheTag := cache_service.Tag{ID: t.ID}
	key := cacheTag.GetTagKey()
	if gredis.Exists(key) {
		data, err := gredis.Get(key)
		if err != nil {
			logging.Info(err)
		} else {
			json.Unmarshal(data, tag)
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

func (t *Tag) GetWithNoCache() (*models.Tag, error) {
	tag, err := models.GetTag(t.ID)
	if err != nil {
		return nil, err
	}

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

func (t *Tag) Export() (string, error) {
	tags, err := t.GetAll()
	if err != nil {
		return "", err
	}

	xlsxFile := xlsx.NewFile()
	sheet, err := xlsxFile.AddSheet("标签信息")
	if err != nil {
		return "", err
	}

	titles := []string{"ID", "名称", "创建人", "创建时间", "修改人", "修改时间"}
	row := sheet.AddRow()

	var cell *xlsx.Cell
	for _, title := range titles {
		cell = row.AddCell()
		cell.Value = title
	}

	for _, v := range tags {
		values := []string{
			strconv.Itoa(v.ID),
			v.Name,
			v.CreatedBy,
			strconv.Itoa(v.CreatedOn),
			v.ModifiedBy,
			strconv.Itoa(v.ModifiedOn),
		}

		row = sheet.AddRow()
		for _, value := range values {
			cell = row.AddCell()
			cell.Value = value
		}
	}

	filePath := export.GetExcelFullPath()
	if err = file.CheckDir(filePath); err != nil {
		return "", err
	}

	time := strconv.Itoa(int(time.Now().Unix()))
	filename := "tags-" + time + ".xlsx"

	if err = xlsxFile.Save(filePath + filename); err != nil {
		return "", err
	}

	return filename, nil
}

func (t *Tag) Import(r io.Reader) error {
	xlsx, err := excelize.OpenReader(r)
	if err != nil {
		return err
	}

	// TODO: 判断 Excel 表格行列格式是否符合规范（避免插入脏数据）
	// TODO: 判断名称是否存在重复（去重）
	rows := xlsx.GetRows("标签信息")
	for irow, row := range rows {
		if irow > 0 {
			tag := map[string]interface{}{
				"name":       row[1],
				"state":      1,
				"created_by": row[2],
			}
			// TODO: 批量插入考虑事务，一旦错误就回滚
			models.AddTag(tag)
		}
	}

	return nil
}
