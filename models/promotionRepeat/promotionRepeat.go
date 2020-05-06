package promotionRepeat

type PromotionRepeat struct {
	Id            uint64 `json:"id"`
	ObjectId      uint64 `json:"object_id"`
	MarketingType uint8  `json:"marketing_type"`
	RepeatType    uint8  `json:"repeat_type"`
	RepeatValue   string `json:"repeat_value"`
	StartHour     string `json:"start_hour"`
	EndHour       string `json:"end_hour"`
}

func GetTableName() string {
	return "shop_promotion_repeat"
}

func GetField() []string {
	return []string{
		"id", "object_id", "marketing_type", "repeat_type", "repeat_value", "start_hour", "end_hour",
	}
}
