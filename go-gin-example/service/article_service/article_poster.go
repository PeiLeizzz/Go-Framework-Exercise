package article_service

import (
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"

	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/qrcode"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/golang/freetype"
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
	BgName string
	*ArticlePoster
	*Rect
	*Pt
}

type Rect struct {
	X0 int
	Y0 int
	X1 int
	Y1 int
}

type Pt struct {
	X int
	Y int
}

func NewArticlePosterBg(bgName string, ap *ArticlePoster, rect *Rect, pt *Pt) *ArticlePosterBg {
	return &ArticlePosterBg{
		BgName:        bgName,
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
		bgF, err := file.Open(path+a.BgName, os.O_RDWR, 0644)
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

		err = a.DrawPoster(&DrawText{
			JPG:    jpg,
			Merged: mergedF,

			Title: "Golang Gin Demo",
			X0:    80,
			Y0:    160,
			Size0: 42,

			SubTitle: "---PeiLei",
			X1:       320,
			Y1:       220,
			Size1:    36,
		}, setting.AppSetting.FontFileName)

		if err != nil {
			file.Delete(a.PosterName, path) // ignore err
			return err
		}
	}

	return nil
}

type DrawText struct {
	JPG    draw.Image
	Merged *os.File

	Title string
	X0    int
	Y0    int
	Size0 float64

	SubTitle string
	X1       int
	Y1       int
	Size1    float64
}

func (a *ArticlePosterBg) DrawPoster(d *DrawText, fontName string) error {
	fontSource := setting.AppSetting.RuntimeRootPath + setting.AppSetting.FontSavePath + fontName
	fontSourceBytes, err := ioutil.ReadFile(fontSource)
	if err != nil {
		return err
	}

	trueTypeFont, err := freetype.ParseFont(fontSourceBytes)
	if err != nil {
		return err
	}

	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(trueTypeFont)
	fc.SetFontSize(d.Size0)
	// 设置剪裁矩形进行绘制
	fc.SetClip(d.JPG.Bounds())
	// 目标图像
	fc.SetDst(d.JPG)
	// 源图像，通常为 image.Uniform
	fc.SetSrc(image.White)

	pt := freetype.Pt(d.X0, d.Y0)
	_, err = fc.DrawString(d.Title, pt)
	if err != nil {
		return err
	}

	fc.SetFontSize(d.Size1)
	_, err = fc.DrawString(d.SubTitle, freetype.Pt(d.X1, d.Y1))
	if err != nil {
		return err
	}

	err = jpeg.Encode(d.Merged, d.JPG, nil)
	if err != nil {
		return err
	}

	return nil
}
