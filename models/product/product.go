package product

import "promotions/models/promotionTool"

type Product struct {
	ProductId uint64 `json:"product_id"`
}

type PromotionProductInfo struct {
	ProductId     uint64  `json:"product_id"`
	ProductCnName string  `json:"product_name"`
	SalePrice     float64 `json:"sale_price"`
}

type ResponseProductDiscount struct {
	ProductId     uint64  `json:"product_id"`
	TotalDiscount float64 `json:"total_discount"`
	CartPrice     float64 `json:"cart_price"`
	Promotions    []*promotionTool.ResponsePromotionDiscount
}

func GetTableName() string {
	return "shop_products"
}

func GetField() []string {
	return []string{
		"product_id", "product_cn_name", "sale_price",
	}
}
