package v2

import (
	"net/http"

	"github.com/PeiLeizzz/go-gin-example/pkg/app"
	"github.com/PeiLeizzz/go-gin-example/pkg/e"
	"github.com/PeiLeizzz/go-gin-example/pkg/export"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/PeiLeizzz/go-gin-example/service/tag_service"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

// @Summary 获取标签列表
// @Produce json
// @Param name query string false "Name"
// @Param state query int false "State"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags [get]
func GetTags(c *gin.Context) {
	appG := app.Gin{C: c}
	valid := validation.Validation{}

	name := c.Query("name")
	if name != "" {
		valid.MaxSize(name, 100, "name").Message("标签名字最多 100 个字符")
	}

	state := -1
	if arg := c.Query("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")
	}

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARMAS, nil)
		return
	}

	tagService := tag_service.Tag{
		PageNum:  util.GetPage(c),
		PageSize: setting.AppSetting.PageSize,
		Name:     name,
		State:    state,
	}
	tags, err := tagService.GetAll()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_TAGS_FAIL, nil)
		return
	}

	count, err := tagService.Count()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_COUNT_TAG_FAIL, nil)
		return
	}

	data := make(map[string]interface{})
	data["lists"] = tags
	data["total"] = count
	appG.Response(http.StatusOK, e.SUCCESS, data)
}

type AddTagForm struct {
	Name      string `form:"name" valid:"Required;MaxSize(100)"`
	CreatedBy string `form:"created_by" valid:"Required;MaxSize(100)"`
	State     int    `form:"state" valid:"Required;Range(0,1)"`
}

// @Summary 新建文章标签
// @Accept mpfd
// @Produce json
// @Param name formData string true "Name"
// @Param created_by formData string true "CreatedBy"
// @Param state formData int true "State"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags [post]
func AddTag(c *gin.Context) {
	var appG = app.Gin{C: c}
	var form AddTagForm

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	tagService := tag_service.Tag{
		Name:      form.Name,
		CreatedBy: form.CreatedBy,
		State:     form.State,
	}
	exists, err := tagService.ExistByName()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}
	// 注意这里是判断是不是已经存在了
	if exists {
		appG.Response(http.StatusOK, e.ERROR_EXIST_TAG, nil)
		return
	}

	err = tagService.Add()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_TAG_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

type EditTagForm struct {
	ID         int    `form:"id" valid:"Required;Min(1)"`
	Name       string `form:"name" valid:"Required;MaxSize(100)"`
	ModifiedBy string `form:"modified_by" valid:"Required;MaxSize(100)"`
	State      int    `form:"state" valid:"Required;Range(0,1)"`
}

// @Summary 更新指定标签
// @Accept mpfd
// @Produce json
// @Param id path int true "ID"
// @Param name formData string true "Name"
// @Param modified_by formData string true "ModifiedBy"
// @Param state formData int true "State"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags/{id} [put]
func EditTag(c *gin.Context) {
	var appG = app.Gin{C: c}
	var form = EditTagForm{
		ID: com.StrTo(c.Param("id")).MustInt(),
	}

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	tagService := tag_service.Tag{
		ID:         form.ID,
		Name:       form.Name,
		ModifiedBy: form.ModifiedBy,
		State:      form.State,
	}
	exists, err := tagService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}

	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_TAG, nil)
		return
	}

	// TODO: 如果标签名和原来的标签名一样，是不是也要允许更改？（只更改状态）
	tag, err := tagService.GetWithNoCache()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}
	if tag.Name != tagService.Name {
		exists, err = tagService.ExistByName()
		if err != nil {
			appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
			return
		}
		if exists {
			appG.Response(http.StatusOK, e.ERROR_EXIST_TAG, nil)
			return
		}
	}

	err = tagService.Edit()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_EDIT_TAG_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// @Summary 删除指定标签
// @Produce json
// @Param id path int true "ID"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags/{id} [delete]
// TODO: 外键依赖还没考虑，如果删除了一个 tag，引用该 tag 的 article 怎么办？
// 思路 1：先查有没有依赖关系？如果有文章引用了该 tag，不让删
// 思路 2：删除 tag 之前把所有依赖该 tag 的文章中的 tag_id 删除
func DeleteTag(c *gin.Context) {
	appG := app.Gin{C: c}
	valid := validation.Validation{}

	id := com.StrTo(c.Param("id")).MustInt()
	valid.Min(id, 1, "id").Message("标签 ID 必须大于 0")

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARMAS, nil)
		return
	}

	tagService := tag_service.Tag{ID: id}
	exists, err := tagService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}

	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_TAG, nil)
		return
	}

	err = tagService.Delete()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_DELETE_TAG_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// @Summary 导出标签（Excel）
// @Accept mpfd
// @Produce json
// @Param name formData string false "Name"
// @Param state formData int false "State"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags/export [post]
func ExportTag(c *gin.Context) {
	appG := app.Gin{C: c}
	valid := validation.Validation{}

	name := c.PostForm("name")
	if name != "" {
		valid.MaxSize(name, 100, "name").Message("标签名字最多 100 个字符")
	}

	state := -1
	if arg := c.PostForm("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")
	}

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARMAS, nil)
		return
	}

	tagService := tag_service.Tag{
		Name:  name,
		State: state,
	}

	filename, err := tagService.Export()
	if err != nil {
		appG.Response(http.StatusOK, e.ERROR_EXPORT_TAG_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, map[string]string{
		"export_url":      export.GetExcelFullUrl(filename),
		"export_save_url": export.GetExportPath() + filename,
	})
}

// @Summary 导入标签（Excel）
// @Accept mpfd
// @Produce json
// @Param file formData file true "Tags Excel File"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/tags/import [post]
func ImportTag(c *gin.Context) {
	appG := app.Gin{C: c}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		logging.Warn(err)
		appG.Response(http.StatusInternalServerError, e.ERROR, nil)
		return
	}

	tagService := tag_service.Tag{}
	err = tagService.Import(file)
	if err != nil {
		logging.Warn(err)
		appG.Response(http.StatusOK, e.ERROR_IMPORT_TAG_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}
