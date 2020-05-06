package services

import (
	"errors"
	"promotions/models/users"
	"promotions/packages/connection"
	"promotions/packages/tools"
	"strings"
)

/**
登录
*/
func Login(username string, password string) (userInfo users.LoginUserResponse, err error) {
	userInfo = users.LoginUserResponse{}
	err = connection.Db.Table(users.GetTableName()).Select(users.GetLoginField()).
		Where("username = ?", username).
		Find(&userInfo).Error
	if err != nil {
		return userInfo, errors.New("账户信息不存在！")
	}
	if strings.Compare(password, userInfo.Password) != 0 {
		return userInfo, errors.New("用户名或密码错误！")
	}

	user := tools.UserInfo{
		UserId:   userInfo.Id,
		Username: userInfo.Username,
		Email:    userInfo.Email,
	}
	result, token := tools.IssueAuthToken(user)
	if !result {
		return userInfo, errors.New("登录失败！")
	}
	userInfo.AccessToken = token

	return userInfo, nil
}
