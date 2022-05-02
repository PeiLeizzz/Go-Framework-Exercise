package article_service

import (
	"image"
	"image/draw"
	"image/jpeg"
	"os"

	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/qrcode"
)

type ArticlePoster struct {
	PosterName string
	*Article
	Qr *qrcode.QrCode
}

func NewArticlePoster(posterName string, article *Article, qr *qrcode.QrCode) *ArticlePoster {
	return &ArticlePoster{
		PosterName: posterName,
		Article:    article,
		Qr:         qr,
	}
}

func GetPosterFlag() string {
	return "poster"
}

func (a *ArticlePoster) CheckMergedImage(path string) bool {
	return !file.CheckNotExist(path + a.PosterName)
}

type ArticlePosterBg struct {
	Name string
	*ArticlePoster
	*Rect
	*Pt
}

type Rect struct {
	Name string
	X0   int
	Y0   int
	X1   int
	Y1   int
}

type Pt struct {
	X int
	Y int
}

func NewArticlePosterBg(name string, ap *ArticlePoster, rect *Rect, pt *Pt) *ArticlePosterBg {
	return &ArticlePosterBg{
		Name:          name,
		ArticlePoster: ap,
		Rect:          rect,
		Pt:            pt,
	}
}

func (a *ArticlePosterBg) Generate(path string) error {
	qrImageFileName, err := a.Qr.Encode(path)
	if err != nil {
		return err
	}

	// 检查图像是否已经存在
	if !a.CheckMergedImage(path) {
		// 打开背景图
		bgF, err := file.Open(path+a.Name, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer bgF.Close()

		// 打开二维码图
		qrF, err := file.Open(path+qrImageFileName, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		defer qrF.Close()

		// 生成待合成的图像
		mergedF, err := file.MustOpen(a.PosterName, path)
		if err != nil {
			return err
		}
		defer mergedF.Close()

		bgImage, err := jpeg.Decode(bgF)
		if err != nil {
			file.Delete(a.PosterName, path) // ignore err
			return err
		}

		qrImage, err := jpeg.Decode(qrF)
		if err != nil {
			file.Delete(a.PosterName, path) // ignore err
			return err
		}

		// 创建新的 RGBA 图像
		jpg := image.NewRGBA(image.Rect(a.Rect.X0, a.Rect.Y0, a.Rect.X1, a.Rect.Y1))

		// 绘制背景
		draw.Draw(jpg, jpg.Bounds(), bgImage, bgImage.Bounds().Min, draw.Over)
		// 在指定 Point 上绘制二维码
		draw.Draw(jpg, jpg.Bounds(), qrImage, qrImage.Bounds().Min.Sub(image.Pt(a.Pt.X, a.Pt.Y)), draw.Over)

		// 将合并的图像以 JPEG 4：2：0 基线格式写入文件
		jpeg.Encode(mergedF, jpg, nil)
	}

	return nil
}
