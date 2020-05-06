package promotionPlatform

type PromotionPlatform struct {
	Id            uint64 `json:"id"`
	ObjectId      uint64 `json:"object_id"`
	MarketingType uint8  `json:"marketing_type"`
	PlatformId    uint64 `json:"platform_id"`
	PlatformType  uint8  `json:"platform_type"`
	PlatformKey   string `json:"platform_key"`
}

func GetTableName() string {
	return "shop_promotion_platform"
}

func GetField() []string {
	return []string{
		"id", "object_id", "marketing_type", "platform_id", "platform_type", "platform_key",
	}
}
