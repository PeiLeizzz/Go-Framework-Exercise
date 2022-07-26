package services

import (
	"errors"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/repositories"
	"golang.org/x/crypto/bcrypt"
)

type IUserService interface {
	IsPwdSuccess(userName, pwd string) (int64, bool)
	InsertUser(*datamodels.User) (int64, error)
}

type UserService struct {
	UserRepository repositories.IUserRepository
}

var _ IUserService = (*UserService)(nil)

func NewUserService(repository repositories.IUserRepository) IUserService {
	return &UserService{
		UserRepository: repository,
	}
}

func ValidatePassword(password, hashed string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return false, errors.New("wrong password!")
	}
	return true, nil
}

func (u *UserService) IsPwdSuccess(userName, pwd string) (int64, bool) {
	user, err := u.UserRepository.SelectByUserName(userName)
	if err != nil {
		return 0, false
	}

	ok, _ := ValidatePassword(pwd, user.HashPassword)
	if !ok {
		return 0, false
	}

	return user.ID, true
}

func GeneratePassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (u *UserService) InsertUser(user *datamodels.User) (int64, error) {
	pwdByte, err := GeneratePassword(user.HashPassword)
	if err != nil {
		return 0, err
	}

	user.HashPassword = string(pwdByte)
	return u.UserRepository.Insert(user)
}
