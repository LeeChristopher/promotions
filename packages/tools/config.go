package tools

var (
	AppConfig *Config
)

type Config struct {
	HttpPort   string
	RunMode    string
	LimitValue int64
	SecretKey  string
}
