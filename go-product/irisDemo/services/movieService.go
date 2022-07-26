package services

import (
	"go-product/irisDemo/repositories"
)

type MovieService interface {
	ShowMovieName() string
}

type MovieServiceManager struct {
	repo repositories.MovieRepository
}

var _ MovieService = (*MovieServiceManager)(nil)

func NewMovieServiceManager(repo repositories.MovieRepository) MovieService {
	return &MovieServiceManager{
		repo: repo,
	}
}

func (m *MovieServiceManager) ShowMovieName() string {
	return "我们获取到的视频名称为：" + m.repo.GetMovieName()
}
