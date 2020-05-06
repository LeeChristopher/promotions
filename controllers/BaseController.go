package controllers

import (
	"promotions/models/users"
	"promotions/packages/connection"
	"promotions/packages/tools"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/go-redis/redis_rate"

	"github.com/astaxie/beego"
)

type InitialiseInterface interface {
	Initialise()
}

type BaseController struct {
	beego.Controller
	StartTime time.Time
}

func (m *BaseController) Prepare() {
	m.StartTime = tools.GetNow()
	if !tools.GetIsSign(m.Ctx.Request.RequestURI) {
		authorization := m.Ctx.Request.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			m.SetResponse(tools.CodeMap["fail"], "Please Submit Certification Information", nil)
			return
		}
		authTokenSlice := strings.Split(authorization, " ")
		if len(authTokenSlice) != 2 || authTokenSlice[0] != "Bearer" {
			m.SetResponse(tools.CodeMap["fail"], "Wrong Authentication Code Format", nil)
			return
		}
		tokenObj, err := jwt.ParseWithClaims(authTokenSlice[1], &tools.AuthToken{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(tools.AppConfig.SecretKey), nil
		})
		if err != nil {
			if !strings.Contains(err.Error(), "expired") {
				m.SetResponse(tools.CodeMap["fail"], "Invalid Authentication Code", nil)
				return
			}
			m.SetResponse(tools.CodeMap["fail"], "Authentication Code Is Expired", nil)
			return
		}
		if tokenObj == nil || !tokenObj.Valid {
			m.SetResponse(tools.CodeMap["fail"], "Invalid Authentication Code", nil)
			return
		}
		claims, ok := tokenObj.Claims.(*tools.AuthToken)
		if !ok {
			m.SetResponse(tools.CodeMap["fail"], "Invalid Authentication Code", nil)
			return
		}

		//限流
		limiter := redis_rate.NewLimiter(connection.Limiter)
		_, _, allow := limiter.Allow("user:limiter:"+strconv.FormatUint(claims.UserId, 10), tools.AppConfig.LimitValue, time.Minute)
		if !allow {
			m.SetResponse(tools.CodeMap["fail"], "Request Too Fast", nil)
		}

		//存储用户信息
		users.LoginUserInfo = &users.User{
			Id:       claims.UserId,
			Username: claims.Username,
			Password: "",
			Email:    claims.Email,
		}
	}

	if app, ok := m.AppController.(InitialiseInterface); ok {
		app.Initialise()
	}
}

func (m *BaseController) SetResponse(code int, message string, data interface{}) {
	response := tools.CommonResponse{
		RunTime: time.Since(m.StartTime).Seconds(),
		Code:    code,
		Message: message,
		Data:    data,
	}

	m.Data["json"] = response
	m.ServeJSON()
}
