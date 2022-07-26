package repositories

import "go-product/irisDemo/datamodels"

type MovieRepository interface {
	GetMovieName() string
}

type MovieManager struct {
}

var _ MovieRepository = (*MovieManager)(nil)

func NewMovieManager() MovieRepository {
	return &MovieManager{}
}

func (m *MovieManager) GetMovieName() string {
	// 模拟赋值给模型
	movie := &datamodels.Movie{
		Name: "peilei-movie",
	}
	return movie.Name
}
