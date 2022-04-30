package v1

import (
	"net/http"

	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/PeiLeizzz/go-gin-example/pkg/e"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

/*
获取标签列表 GET("/tags?name=&state=&page=")
Produce  json
Param name query string false "Name"
Param state query int false "State"
Success 200 {object} gin.H "{"code":200,"data":{},"msg":"ok"}"
Router /api/v1/tags [get]
*/
func GetTags(c *gin.Context) {
	name := c.Query("name")

	data := make(map[string]interface{})
	maps := map[string]interface{}{"deleted_on": 0}
	valid := validation.Validation{}

	if name != "" {
		maps["name"] = name

		valid.MaxSize(name, 100, "name").Message("标签名字最多 100 个字符")
	}

	if arg := c.Query("state"); arg != "" {
		state := com.StrTo(arg).MustInt()
		maps["state"] = state

		valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")
	}

	code := e.INVALID_PARMAS
	if !valid.HasErrors() {
		code = e.SUCCESS
		data["lists"] = models.GetTags(util.GetPage(c), setting.AppSetting.PageSize, maps)
		data["total"] = models.GetTagTotal(maps)
	} else {
		for _, err := range valid.Errors {
			logging.Info(err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}

/*
Summary 新建文章标签
Produce  json
Param name query string true "Name"
Param state query int false "State"
Param created_by query int true "CreatedBy"
Success 200 {object} gin.H "{"code":200,"data":{},"msg":"ok"}"
Router /api/v1/tags [post]
*/
func AddTag(c *gin.Context) {
	name := c.Query("name")
	state := com.StrTo(c.DefaultQuery("state", "0")).MustInt()
	createdBy := c.Query("created_by")

	valid := validation.Validation{}
	valid.Required(name, "name").Message("名字不能为空")
	valid.MaxSize(name, 100, "name").Message("名称最长为 100 字符")
	valid.Required(createdBy, "created_by").Message("创建人不能为空")
	valid.MaxSize(createdBy, 100, "created_by").Message("创建人最长为 100 字符")
	valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")

	code := e.INVALID_PARMAS
	if !valid.HasErrors() {
		if !models.ExistTagByName(name) {
			code = e.SUCCESS
			models.AddTag(name, state, createdBy)
		} else {
			code = e.ERROR_EXIST_TAG
		}
	} else {
		for _, err := range valid.Errors {
			logging.Info(err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": make(map[string]string),
	})
}

/*
Summary 更新指定标签
Produce  json
Param id path int true "ID"
Param name query string true "Name"
Param state query int false "State"
Param modified_by query string true "ModifiedBy"
Success 200 {object} gin.H "{"code":200,"data":{},"msg":"ok"}"
Router /api/v1/tags/{id} [put]
*/
func EditTag(c *gin.Context) {
	id := com.StrTo(c.Param("id")).MustInt()
	name := c.Query("name")
	modifiedBy := c.Query("modified_by")

	valid := validation.Validation{}

	state := -1
	if arg := c.Query("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")
	}

	valid.Required(id, "id").Message("ID 不能为空")
	valid.Required(modifiedBy, "modified_by").Message("修改人不能为空")
	valid.MaxSize(modifiedBy, 100, "modified_by").Message("修改人最长为 100 字符")
	valid.Required(name, "name").Message("名称不能为空")
	valid.MaxSize(name, 100, "name").Message("名称最长为 100 字符")

	code := e.INVALID_PARMAS
	if !valid.HasErrors() {
		code = e.SUCCESS
		if models.ExistTagByID(id) {
			data := make(map[string]interface{})
			data["modified_by"] = modifiedBy
			if name != "" {
				data["name"] = name
			}
			if state != -1 {
				data["state"] = state
			}
			models.EditTag(id, data)
		} else {
			code = e.ERROR_NOT_EXIST_TAG
		}
	} else {
		for _, err := range valid.Errors {
			logging.Info(err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": make(map[string]string),
	})
}

/*
Summary 删除指定标签
Produce  json
Param id path int true "ID"
Success 200 {object} gin.H "{"code":200,"data":{},"msg":"ok"}"
Router /api/v1/tags/{id} [delete]
*/
func DeleteTag(c *gin.Context) {
	id := com.StrTo(c.Param("id")).MustInt()

	valid := validation.Validation{}
	valid.Min(id, 1, "id").Message("ID 必须大于 0")

	code := e.INVALID_PARMAS
	if !valid.HasErrors() {
		code = e.SUCCESS
		if models.ExistTagByID(id) {
			models.DeleteTag(id)
		} else {
			code = e.ERROR_NOT_EXIST_TAG
		}
	} else {
		for _, err := range valid.Errors {
			logging.Info(err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": make(map[string]string),
	})
}
