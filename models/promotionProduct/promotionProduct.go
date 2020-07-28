package promotionProduct

type PromotionProduct struct {
	Id             uint64  `json:"id"`
	PromotionalId  uint64  `json:"promotional_id"`
	IsAdd          uint8   `json:"is_add"`
	ProductId      uint64  `json:"product_id"`
	StockLimit     uint8   `json:"stock_limit"`
	Stock          uint64  `json:"stock"`
	PromotionPrice float64 `json:"promotion_price"`
	Discount       float64 `json:"discount"`
	ChannelId      uint64  `json:"channel_id"`
}

type RequestPromotionBaseProduct struct {
	ProductId uint64  `json:"product_id"`
	Price     float64 `json:"price"`
	Quantity  uint64  `json:"quantity"`
}

type RequestPromotionProduct struct {
	RequestPromotionBaseProduct
	ProductType   string                         `json:"product_type"`
	IsSelected    uint8                          `json:"is_selected"`
	ProductDetail []*RequestPromotionBaseProduct `json:"product_detail"`
}

type PromotionStock struct {
	Stock      int64  `json:"stock"`
	StockLimit uint8  `json:"stock_limit"`
	Success    uint8  `json:"success"`
	ProductId  uint64 `json:"product_id"`
}

func GetTableName() string {
	return "shop_promotional_products"
}

func GetField() []string {
	return []string{
		"id", "promotional_id", "is_add", "product_id", "stock_limit", "stock", "promotion_price", "discount", "channel_id",
	}
}
