package v2

import (
	"net/http"

	"github.com/PeiLeizzz/go-gin-example/pkg/app"
	"github.com/PeiLeizzz/go-gin-example/pkg/e"
	"github.com/PeiLeizzz/go-gin-example/pkg/qrcode"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/PeiLeizzz/go-gin-example/service/article_service"
	"github.com/PeiLeizzz/go-gin-example/service/tag_service"
	"github.com/astaxie/beego/validation"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

// @Summary 获取文章列表
// @Produce  json
// @Param tag_id query int false "TagID"
// @Param state query int false "State"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles [get]
func GetArticles(c *gin.Context) {
	appG := app.Gin{C: c}
	valid := validation.Validation{}

	state := -1
	if arg := c.Query("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state").Message("状态只允许 0 或 1")
	}

	tagId := -1
	if arg := c.Query("tag_id"); arg != "" {
		tagId = com.StrTo(arg).MustInt()
		valid.Min(tagId, 1, "tag_id").Message("标签 ID 必须大于 0")
	}

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARMAS, nil)
		return
	}

	articleService := article_service.Article{
		TagID:    tagId,
		State:    state,
		PageNum:  util.GetPage(c),
		PageSize: setting.AppSetting.PageSize,
	}

	articles, err := articleService.GetAll()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_ARTICLES_FAIL, nil)
		return
	}

	count, err := articleService.Count()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_COUNT_ARTICLE_FAIL, nil)
		return
	}

	data := make(map[string]interface{})
	data["lists"] = articles
	data["total"] = count

	appG.Response(http.StatusOK, e.SUCCESS, data)
}

// @Summary 获取指定文章
// @Produce  json
// @Param id path int true "ID"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles/{id} [get]
func GetArticle(c *gin.Context) {
	appG := app.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	valid := validation.Validation{}
	valid.Min(id, 1, "id")

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARMAS, nil)
		return
	}

	articleService := article_service.Article{ID: id}
	exists, err := articleService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_ARTICLE_FAIL, nil)
		return
	}
	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_ARTICLE, nil)
		return
	}

	article, err := articleService.Get()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_ARTICLE_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, article)
}

type AddArticleForm struct {
	TagID         int    `form:"tag_id" valid:"Required;Min(1)"`
	Title         string `form:"title" valid:"Required;MaxSize(100)"`
	Desc          string `form:"desc" valid:"Required;MaxSize(255)"`
	Content       string `form:"content" valid:"Required;MaxSize(65535)"`
	CreatedBy     string `form:"created_by" valid:"Required;MaxSize(100)"`
	CoverImageUrl string `form:"cover_image_url" valid:"MaxSize(255)"`
	State         int    `form:"state" valid:"Required;Range(0,1)"`
}

// @Summary 新建文章
// @Accept mpfd
// @Produce  json
// @Param tag_id formData int true "TagID"
// @Param title formData string true "Title"
// @Param desc formData string true "Desc"
// @Param content formData string true "Content"
// @Param created_by formData string true "CreatedBy"
// @Param state formData int true "State"
// @Param cover_image_url formData string false "CoverImageUrl"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles [post]
func AddArticle(c *gin.Context) {
	var appG = app.Gin{C: c}
	var form AddArticleForm

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	tagService := tag_service.Tag{ID: form.TagID}
	exists, err := tagService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}

	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_TAG, nil)
		return
	}

	articleService := article_service.Article{
		TagID:         form.TagID,
		Title:         form.Title,
		Desc:          form.Desc,
		Content:       form.Content,
		CoverImageUrl: form.CoverImageUrl,
		State:         form.State,
		CreateBy:      form.CreatedBy,
	}

	if err := articleService.Add(); err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_ADD_ARTICLE_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

type EditArticleForm struct {
	ID            int    `form:"id" valid:"Required;Min(1)"`
	TagID         int    `form:"tag_id" valid:"Required;Min(1)"`
	Title         string `form:"title" valid:"Required;MaxSize(100)"`
	Desc          string `form:"desc" valid:"Required;MaxSize(255)"`
	Content       string `form:"content" valid:"Required;MaxSize(65535)"`
	ModifiedBy    string `form:"modified_by" valid:"Required;MaxSize(100)"`
	CoverImageUrl string `form:"cover_image_url" valid:"MaxSize(255)"`
	State         int    `form:"state" valid:"Required;Range(0,1)"`
}

// @Summary 更新指定文章
// @Accept mpfd
// @Produce  json
// @Param id path int true "ID"
// @Param tag_id formData string true "TagID"
// @Param title formData string true "Title"
// @Param desc formData string true "Desc"
// @Param content formData string true "Content"
// @Param modified_by formData string true "ModifiedBy"
// @Param state formData int true "State"
// @Param cover_image_url formData string false "CoverImageUrl"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles/{id} [put]
func EditArticle(c *gin.Context) {
	var appG = app.Gin{C: c}
	var form = EditArticleForm{
		ID:            com.StrTo(c.Param("id")).MustInt(),
		CoverImageUrl: "",
	}

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	// TODO: 这里如果 state 不在 body 中，默认为 0
	// 可能导致，本不想修改 state，却将其修改为了 0
	// 在 Tag 中也有这个问题
	// 可能的解决办法：1. 修改数据库 1 为启用，2 为禁用
	//				2. 能否给 state 加一个默认值？
	//				3. 先把 article 查出来，再更新结构体，再 Edit？
	articleService := article_service.Article{
		ID:         form.ID,
		TagID:      form.TagID,
		Title:      form.Title,
		Desc:       form.Desc,
		Content:    form.Content,
		ModifiedBy: form.ModifiedBy,
		State:      form.State,
	}
	if form.CoverImageUrl != "" {
		articleService.CoverImageUrl = form.CoverImageUrl
	}
	exists, err := articleService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_ARTICLE_FAIL, nil)
		return
	}

	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_ARTICLE, nil)
		return
	}

	tagService := tag_service.Tag{ID: form.TagID}
	exists, err = tagService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_TAG_FAIL, nil)
		return
	}

	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_TAG, nil)
		return
	}

	err = articleService.Edit()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_EDIT_ARTICLE_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

// @Summary 删除指定文章
// @Produce  json
// @Param id path int true "ID"
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles/{id} [delete]
func DeleteArticle(c *gin.Context) {
	appG := app.Gin{C: c}
	id := com.StrTo(c.Param("id")).MustInt()

	valid := validation.Validation{}
	valid.Min(id, 1, "id")

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusOK, e.INVALID_PARMAS, nil)
		return
	}

	articleService := article_service.Article{ID: id}
	exists, err := articleService.ExistByID()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_CHECK_EXIST_ARTICLE_FAIL, nil)
		return
	}
	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_ARTICLE, nil)
		return
	}

	err = articleService.Delete()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_DELETE_ARTICLE_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, nil)
}

const (
	// TODO: 这里之后需要改为文章 id 对应的前端路径
	QRCODE_URL                = "https://github.com/EDDYCJY/blog#gin%E7%B3%BB%E5%88%97%E7%9B%AE%E5%BD%95"
	BACKGROUND_IMAGE_FILENAME = "bg.jpg"
)

// @Summary 生成文章二维码
// @Accept mpfd
// @Produce  json
// @Security x-token
// @param x-token header string true "Authorization"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /api/v2/articles/poster/generate [post]
// TODO: 这里之后要有个 path {id} 参数，得到对应的文章链接，再传入 NewQrCode
func GenerateArticlePoster(c *gin.Context) {
	appG := app.Gin{C: c}

	article := &article_service.Article{}                        // 之后改为具体的文章
	qrc := qrcode.NewQrCode(QRCODE_URL, 300, 300, qr.M, qr.Auto) // 具体文章的路径

	posterName := article_service.GetPosterFlag() + "-" + qrcode.GetQrCodeFileName(qrc.URL) + qrc.GetQrCodeExt()

	articlePosterBgService := article_service.NewArticlePosterBg(
		BACKGROUND_IMAGE_FILENAME,
		article_service.NewArticlePoster(posterName, article, qrc),
		&article_service.Rect{
			X0: 0,
			Y0: 0,
			X1: 550,
			Y1: 700,
		},
		&article_service.Pt{
			X: 125,
			Y: 298,
		},
	)

	path := qrcode.GetQrCodeFullPath()
	if err := articlePosterBgService.Generate(path); err != nil {
		appG.Response(http.StatusOK, e.ERROR_GEN_ARTICLE_POSTER_FAIL, nil)
		return
	}

	data := map[string]string{
		"poster_url":      qrcode.GetQrCodeFullUrl(posterName),
		"poster_save_url": path + posterName,
	}
	appG.Response(http.StatusOK, e.SUCCESS, data)
}

// TODO: 文章的导入导出
