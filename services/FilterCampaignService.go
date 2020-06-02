package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"promotions/models/product"
	"promotions/models/promotionMember"
	"promotions/models/promotionPlatform"
	"promotions/models/promotionProduct"
	"promotions/models/promotionRepeat"
	"promotions/models/promotionTool"
	"promotions/models/selfProduct"
	"promotions/packages/connection"
	"promotions/packages/tools"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
)

type ResponseDiscountList struct {
	TotalDiscount     float64                                `json:"total_discount"`
	PromotionDiscount float64                                `json:"promotion_discount"`
	CouponDiscount    float64                                `json:"coupon_discount"`
	ProductDiscount   []*product.ResponseProductDiscount     `json:"product_discount"`
	PromotionList     []*promotionTool.ResponsePromotionList `json:"promotion_list"`
	Freight           float64                                `json:"freight"`
	ShouldPayment     float64                                `json:"should_payment"`
}

func GetValidCampaign(businessId uint64, memberId uint64, isNewMember uint8) (promotionToolList []*promotionTool.PromotionTool, promotionToolIdList []uint64, err error) {
	nowTimeUnix := tools.GetNow().Unix()

	promotionToolList = make([]*promotionTool.PromotionTool, 0, 32)
	query := connection.Db.Table(promotionTool.GetTableName()).Select(promotionTool.GetField()).
		Where("start_time <= ? AND end_time >= ? AND marketing_platform = ? AND business_id = ? AND status = ?", nowTimeUnix, nowTimeUnix, 1, businessId, 2)
	if memberId == 0 {
		query = query.Where("marketing_member = ?", 1)
	}
	if isNewMember == 2 {
		query = query.Where("limit_new = ?", 2)
	}
	err = query.Find(&promotionToolList).Error
	if err != nil {
		return promotionToolList, nil, err
	}

	promotionToolListLen := len(promotionToolList)
	promotionToolIdList = make([]uint64, 0, promotionToolListLen)
	repeatPromotionIdList := make([]uint64, 0, promotionToolListLen)
	for i := range promotionToolList {
		if promotionToolList[i].IsRepeat == 2 {
			repeatPromotionIdList = append(repeatPromotionIdList, promotionToolList[i].PromotionalId)
			continue
		}
		promotionToolIdList = append(promotionToolIdList, promotionToolList[i].PromotionalId) //不重复的活动id
	}
	repeatPromotionIdListLen := len(repeatPromotionIdList)
	if repeatPromotionIdListLen == 0 {
		return promotionToolList, promotionToolIdList, nil
	}
	validRepeatPromotionIdList, err := getSmallTimeValidCampaign(repeatPromotionIdList, repeatPromotionIdListLen)
	if err != nil {
		return nil, nil, err
	}
	if len(validRepeatPromotionIdList) > 0 {
		validPromotionList := make([]*promotionTool.PromotionTool, 0, promotionToolListLen)

		for i := range promotionToolList {
			if promotionToolList[i].IsRepeat == 1 {
				validPromotionList = append(validPromotionList, promotionToolList[i])
			}
			if !tools.InUint64(promotionToolList[i].PromotionalId, validRepeatPromotionIdList) {
				continue
			}
			validPromotionList = append(validPromotionList, promotionToolList[i])
			promotionToolIdList = append(promotionToolIdList, promotionToolList[i].PromotionalId)
		}

		promotionToolList = validPromotionList
	}

	return promotionToolList, promotionToolIdList, nil
}

func getSmallTimeValidCampaign(repeatPromotionIdList []uint64, repeatPromotionIdListLen int) (validRepeatPromotionIdList []uint64, err error) {
	promotionRepeatList := make([]*promotionRepeat.PromotionRepeat, 0, repeatPromotionIdListLen)
	err = connection.Db.Table(promotionRepeat.GetTableName()).Select(promotionRepeat.GetField()).
		Where("object_id in (?) AND marketing_type = ?", repeatPromotionIdList, 1).
		Find(&promotionRepeatList).Error
	if err != nil {
		return nil, err
	}

	nowTime := tools.GetNow()
	nowWeekday := strconv.Itoa(int(nowTime.Weekday()))
	nowDay := strconv.Itoa(nowTime.Day())
	nowDate := nowTime.Format(tools.DATA_FORMART)
	location, _ := time.LoadLocation(tools.LOCATION)
	validRepeatPromotionIdList = make([]uint64, 0, repeatPromotionIdListLen)
	for i := range promotionRepeatList {
		startTime, err := time.ParseInLocation(tools.DATATIME_FORMART, fmt.Sprintf("%s %s", nowDate, promotionRepeatList[i].StartHour), location)
		if err != nil {
			return nil, err
		}
		endTime, err := time.ParseInLocation(tools.DATATIME_FORMART, fmt.Sprintf("%s %s", nowDate, promotionRepeatList[i].EndHour), location)
		if err != nil {
			return nil, err
		}
		if startTime.After(nowTime) || endTime.Before(nowTime) {
			continue
		}

		repeatValue := make([]string, 0, 32)
		err = json.Unmarshal([]byte(promotionRepeatList[i].RepeatValue), &repeatValue)
		if err != nil {
			continue
		}
		if promotionRepeatList[i].RepeatType == 2 {
			if !tools.InString(nowWeekday, repeatValue) {
				continue
			}
		}
		if promotionRepeatList[i].RepeatType == 3 {
			if !tools.InString(nowDay, repeatValue) {
				continue
			}
		}
		validRepeatPromotionIdList = append(validRepeatPromotionIdList, promotionRepeatList[i].ObjectId)
	}

	return validRepeatPromotionIdList, nil
}

func GetValidMemberLevel(promotionToolList *[]*promotionTool.PromotionTool, promotionToolIdList *[]uint64, memberId uint64) (err error) {
	promotionToolListLen := len(*promotionToolIdList)
	validPromotionList := make([]*promotionTool.PromotionTool, 0, promotionToolListLen)
	validPromotionIdList := make([]uint64, 0, promotionToolListLen)
	defer func() {
		*promotionToolList = validPromotionList
		*promotionToolIdList = validPromotionIdList
	}()
	if promotionToolListLen == 0 {
		return
	}

	needMemberLevelCampaignIdList := make([]uint64, 0, promotionToolListLen)
	for i := range *promotionToolList {
		if (*promotionToolList)[i].MarketingMember == 2 {
			needMemberLevelCampaignIdList = append(needMemberLevelCampaignIdList, (*promotionToolList)[i].PromotionalId)
			continue
		}

		validPromotionList = append(validPromotionList, (*promotionToolList)[i])
		validPromotionIdList = append(validPromotionIdList, (*promotionToolList)[i].PromotionalId)
	}
	if len(needMemberLevelCampaignIdList) == 0 || memberId == 0 {
		return
	}

	var userLevel uint32 = 1
	promotionMemberList := make([]promotionMember.PromotionMember, 0, 32)
	err = connection.Db.Table(promotionMember.GetTableName()).Select(promotionMember.GetField()).
		Where("object_id in (?) marketing_type = ?", needMemberLevelCampaignIdList, 1).
		Find(&promotionMemberList).Error
	if err != nil {
		return errors.New("出错了！")
	}
	promotionIdLevelMap := make(map[uint64][]uint32, len(promotionMemberList))
	for i := range promotionMemberList {
		if _, ok := promotionIdLevelMap[promotionMemberList[i].ObjectId]; ok {
			promotionIdLevelMap[promotionMemberList[i].ObjectId] = append(promotionIdLevelMap[promotionMemberList[i].ObjectId], promotionMemberList[i].MemberRankLevel)
		} else {
			levelIdList := make([]uint32, 0, 8)
			levelIdList = append(levelIdList, promotionMemberList[i].MemberRankLevel)
			promotionIdLevelMap[promotionMemberList[i].ObjectId] = levelIdList
		}
	}

	for i := range *promotionToolList {
		if (*promotionToolList)[i].MarketingMember == 1 {
			continue
		}
		levelIdList, ok := promotionIdLevelMap[(*promotionToolList)[i].PromotionalId]
		if !ok {
			continue
		}
		if !tools.InUint32(userLevel, levelIdList) {
			continue
		}
		validPromotionList = append(validPromotionList, (*promotionToolList)[i])
		validPromotionIdList = append(validPromotionIdList, (*promotionToolList)[i].PromotionalId)
	}

	return nil
}

func GetValidPlatform(promotionToolList *[]*promotionTool.PromotionTool, promotionToolIdList *[]uint64, platform string) (err error) {
	promotionToolLen := len(*promotionToolIdList)
	promotionPlatformList := make([]*promotionPlatform.PromotionPlatform, 0, promotionToolLen)
	err = connection.Db.Table(promotionPlatform.GetTableName()).Select(promotionPlatform.GetField()).
		Where("object_id in (?) AND marketing_type = ? AND platform_type = ?", *promotionToolIdList, 1, 1).
		Find(&promotionPlatformList).Error
	if err != nil {
		return errors.New("出错了！")
	}

	validPromotionIdList := make([]uint64, 0, promotionToolLen)
	validPromotionList := make([]*promotionTool.PromotionTool, 0, promotionToolLen)
	defer func() {
		*promotionToolList = validPromotionList
		*promotionToolIdList = validPromotionIdList
	}()

	promotionIdPlatformMap := make(map[uint64][]string, promotionToolLen)
	for i := range promotionPlatformList {
		if _, ok := promotionIdPlatformMap[promotionPlatformList[i].ObjectId]; ok {
			promotionIdPlatformMap[promotionPlatformList[i].ObjectId] = append(promotionIdPlatformMap[promotionPlatformList[i].ObjectId], promotionPlatformList[i].PlatformKey)
		} else {
			platformList := make([]string, 0, 8)
			platformList = append(platformList, promotionPlatformList[i].PlatformKey)
			promotionIdPlatformMap[promotionPlatformList[i].ObjectId] = platformList
		}
	}

	for i := range *promotionToolList {
		platformList, ok := promotionIdPlatformMap[(*promotionToolList)[i].PromotionalId]
		if !ok {
			continue
		}
		if !tools.InString(platform, platformList) {
			continue
		}
		validPromotionList = append(validPromotionList, (*promotionToolList)[i])
		validPromotionIdList = append(validPromotionIdList, (*promotionToolList)[i].PromotionalId)
	}

	return nil
}

func GetValidPromotionProduct(promotionToolList *[]*promotionTool.PromotionTool, promotionToolIdList *[]uint64, cartProductList []*promotionProduct.RequestPromotionProduct, campaignProductMap *map[uint64][]*promotionProduct.PromotionProduct, businessId uint64, cartProductInfoMap *map[uint64]*product.PromotionProductInfo) (totalPrice float64, err error) {
	promotionToolLen := len(*promotionToolIdList)
	validPromotionIdList := make([]uint64, 0, promotionToolLen)
	validPromotionList := make([]*promotionTool.PromotionTool, 0, promotionToolLen)
	campaignProductMapList := make(map[uint64][]*promotionProduct.PromotionProduct, promotionToolLen)
	defer func() {
		*promotionToolList = validPromotionList
		*promotionToolIdList = validPromotionIdList
		*campaignProductMap = campaignProductMapList
	}()

	promotionProductList := make([]*promotionProduct.PromotionProduct, 0, promotionToolLen)
	err = connection.Db.Table(promotionProduct.GetTableName()).Select(promotionProduct.GetField()).
		Where("promotional_id in (?) AND channel_id = ?", *promotionToolIdList, 1).
		Find(&promotionProductList).Error
	if err != nil {
		return 0, errors.New("出错了！")
	}

	cartProductIdList := make([]uint64, 0, len(cartProductList))
	relatedPromotionIdList := make([]uint64, 0, promotionToolLen)
	for i := range cartProductList {
		cartProductIdList = append(cartProductIdList, cartProductList[i].ProductId)
	}
	for i := range promotionProductList {
		if !tools.InUint64(promotionProductList[i].ProductId, cartProductIdList) {
			continue
		}
		if _, ok := campaignProductMapList[promotionProductList[i].PromotionalId]; ok {
			campaignProductMapList[promotionProductList[i].PromotionalId] = append(campaignProductMapList[promotionProductList[i].PromotionalId], promotionProductList[i])
		} else {
			promotionProductInfoList := make([]*promotionProduct.PromotionProduct, 0, 16)
			promotionProductInfoList = append(promotionProductInfoList, promotionProductList[i])
			campaignProductMapList[promotionProductList[i].PromotionalId] = promotionProductInfoList
		}
		if tools.InUint64(promotionProductList[i].PromotionalId, relatedPromotionIdList) {
			continue
		}

		relatedPromotionIdList = append(relatedPromotionIdList, promotionProductList[i].PromotionalId)
	}
	for i := range *promotionToolList {
		if !tools.InUint64((*promotionToolList)[i].PromotionalId, relatedPromotionIdList) {
			continue
		}
		validPromotionList = append(validPromotionList, (*promotionToolList)[i])
		validPromotionIdList = append(validPromotionIdList, (*promotionToolList)[i].PromotionalId)
	}

	//获取到商品信息
	cartProductIdListLen := len(cartProductIdList)
	productInfoList := make([]*product.PromotionProductInfo, 0, cartProductIdListLen)
	err = connection.Db.Table(selfProduct.GetTableName()).Select(selfProduct.GetField()).
		Where("business_id = ? AND product_id in (?) AND status = ?", businessId, cartProductIdList, 1).
		Find(&productInfoList).Error
	if err != nil {
		return 0, errors.New("选购商品未上架或商品库不存在！")
	}
	var totalPriceDecimal decimal.Decimal
	for i := range productInfoList {
		(*cartProductInfoMap)[productInfoList[i].ProductId] = productInfoList[i]
		totalPriceDecimal = decimal.NewFromFloat(0).Add(decimal.NewFromFloat(productInfoList[i].SalePrice))
	}
	totalPrice, _ = totalPriceDecimal.Float64()

	return totalPrice, nil
}
