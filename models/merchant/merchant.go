package merchant

type Merchant struct {
	BusinessId   uint64 `json:"business_id"`
	BusinessName string `json:"business_name"`
	BusinessKey  string `json:"business_key"`
	Status       uint8  `json:"status"`
}

func GetTableName() string {
	return "shop_merchants"
}

func GetField() []string {
	return []string{
		"business_id", "business_name", "business_key", "status",
	}
}
