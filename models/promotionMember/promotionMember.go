package promotionMember

type PromotionMember struct {
	Id              uint64 `json:"id"`
	ObjectId        uint64 `json:"object_id"`
	MarketingType   uint8  `json:"marketing_type"`
	MemberRankLevel uint32 `json:"member_rank_level"`
}

func GetTableName() string {
	return "shop_promotion_member"
}

func GetField() []string {
	return []string{
		"id", "object_id", "marketing_type", "member_rank_level",
	}
}
