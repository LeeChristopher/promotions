package selfProduct

type SelfProduct struct {
	Id        uint64 `json:"id"`
	ProductId uint64 `json:"product_id"`
}

func GetTableName() string {
	return "shop_self_products"
}

func GetField() []string {
	return []string{
		"product_id", "channel_cn_name", "sale_price",
	}
}
