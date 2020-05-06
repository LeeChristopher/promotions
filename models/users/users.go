package users

var (
	LoginUserInfo *User
)

type User struct {
	Id       uint64 `json:"-"`
	Username string `json:"username"`
	Password string `json:"-"`
	Email    string `json:"email"`
}

type LoginUserResponse struct {
	User
	AccessToken string `json:"access_token"`
}

func GetTableName() string {
	return "users"
}

func GetLoginField() []string {
	return []string{
		"id", "username", "password",
	}
}
