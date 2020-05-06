package promotionTool

import (
	"promotions/models/promotionProduct"
	"time"
)

type PromotionTool struct {
	PromotionalId           uint64    `json:"promotional_id"`
	PromotionalName         string    `json:"promotional_name"`
	RuleDesc                string    `json:"rule_desc"`
	IsOnline                uint8     `json:"is_online"`
	IsAutoCompute           uint8     `json:"is_auto_compute"`
	IsOnlyMember            uint8     `json:"is_only_member"`
	LimitNew                uint8     `json:"limit_new"`
	LimitBuy                uint8     `json:"limit_buy"`
	LimitValue              uint32    `json:"limit_value"`
	StartTime               uint64    `json:"start_time"`
	EndTime                 uint64    `json:"end_time"`
	IsRepeat                uint8     `json:"is_repeat"`
	PromotionalTypeId       uint64    `json:"promotional_type_id"`
	PromotionalTypeCategory string    `json:"promotional_type_category"`
	Conditions              string    `json:"conditions"`
	PriorityIndex           uint8     `json:"priority_index"`
	IsExclusive             uint8     `json:"is_exclusive"`
	ProductRangeType        uint8     `json:"product_range_type"`
	ProductRangeValue       string    `json:"product_range_value"`
	Status                  uint8     `json:"-"`
	CreatedAt               time.Time `json:"created_at"`
}

type RequestPromotionParam struct {
	BusinessKey string
	BusinessId  uint64
	MemberId    uint64
	Platform    string
	ProductList []*promotionProduct.RequestPromotionProduct
	IsNewMember uint8
	Freight     uint64
	FreightCost uint64
}

type ResponsePromotionDiscount struct {
	PromotionalId   uint64  `json:"promotional_id"`
	PromotionalName string  `json:"promotional_name"`
	Discount        float64 `json:"discount"`
}

type ResponseMatchProduct struct {
	ProductId uint64  `json:"product_id"`
	Discount  float64 `json:"discount"`
}

type ResponsePromotionList struct {
	ResponsePromotionDiscount
	MatchStatus   uint8                   `json:"match_status"`
	MatchProducts []*ResponseMatchProduct `json:"match_products"`
}

type SortPromotionTool []*PromotionTool

func (m SortPromotionTool) Len() int {
	return len(m)
}

func (m SortPromotionTool) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m SortPromotionTool) Less(i, j int) bool {
	if m[i].PriorityIndex > m[j].PriorityIndex {
		return false
	}
	if m[i].PriorityIndex < m[j].PriorityIndex {
		return true
	}
	if m[i].CreatedAt.After(m[j].CreatedAt) {
		return true
	}

	return false
}

func GetTableName() string {
	return "shop_promotional_tools"
}

func GetField() []string {
	return []string{
		"promotional_id", "promotional_name", "rule_desc", "is_online", "is_auto_compute", "is_only_member", "limit_new", "limit_buy", "limit_value", "start_time", "end_time",
		"is_repeat", "promotional_type_id", "promotional_type_category", "conditions", "priority_index", "is_exclusive", "product_range_type", "product_range_value", "status", "created_at",
	}
}
