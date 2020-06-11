package services

import (
	"errors"
	"promotions/models/merchant"
	"promotions/models/product"
	"promotions/models/promotionProduct"
	"promotions/models/promotionTool"
	"promotions/packages/connection"
	"promotions/packages/tools"
	"sort"

	"github.com/shopspring/decimal"
)

type CampaignService struct {
	RequestParam          *promotionTool.RequestPromotionParam
	CampaignProductMap    map[uint64][]*promotionProduct.PromotionProduct
	CartProductInfoMap    map[uint64]*product.PromotionProductInfo
	ResponseDiscountList  *ResponseDiscountList
	TotalDiscount         float64
	PromotionDiscount     float64
	ProductDiscountIdList []uint64
	ProductDiscount       []*product.ResponseProductDiscount
	PromotionList         []*promotionTool.ResponsePromotionList
	TotalPrice            float64
}

func NewCampaign(requestParam *promotionTool.RequestPromotionParam) *CampaignService {
	return &CampaignService{
		RequestParam:          requestParam,
		CampaignProductMap:    make(map[uint64][]*promotionProduct.PromotionProduct, 32),
		CartProductInfoMap:    make(map[uint64]*product.PromotionProductInfo, 32),
		ProductDiscount:       make([]*product.ResponseProductDiscount, 0, 16),
		PromotionList:         make([]*promotionTool.ResponsePromotionList, 0, 16),
		ProductDiscountIdList: make([]uint64, 0, 8),
	}
}

func (m *CampaignService) GetDiscountList() (result *ResponseDiscountList, err error) {
	businessInfo := &merchant.Merchant{}
	err = connection.Db.Table(merchant.GetTableName()).Select("business_id").
		Where("business_key = ? AND status = ?", m.RequestParam.BusinessKey, 1).
		Find(businessInfo).Error
	if err != nil {
		return nil, errors.New("商户未找到！")
	}
	m.RequestParam.BusinessId = businessInfo.BusinessId

	//获取有效时间段内手段
	promotionToolList, promotionToolIdList, err := GetValidCampaign(businessInfo.BusinessId, m.RequestParam.MemberId, m.RequestParam.IsNewMember)
	if err != nil {
		return nil, err
	}

	//验证会员等级
	err = GetValidMemberLevel(&promotionToolList, &promotionToolIdList, m.RequestParam.MemberId)
	if err != nil {
		return nil, err
	}

	//验证平台
	err = GetValidPlatform(&promotionToolList, &promotionToolIdList, m.RequestParam.Platform)
	if err != nil {
		return nil, err
	}

	//验证商品是否是活动商品 获取到的信息:活动下的商品信息  商品信息
	m.TotalPrice, err = GetValidPromotionProduct(&promotionToolList, &promotionToolIdList, m.RequestParam.ProductList, &m.CampaignProductMap, m.RequestParam.BusinessId, &m.CartProductInfoMap)
	if err != nil {
		return nil, err
	}

	//处理响应活动列表
	m.setResponsePromotionList(promotionToolList)

	m.campaignPartitionCompute(promotionToolList, promotionToolIdList)

	result = m.buildUpResponse()

	return result, nil
}

func (m *CampaignService) campaignPartitionCompute(promotionToolList []*promotionTool.PromotionTool, promotionToolIdList []uint64) {
	//对有效的活动排序
	var sortPromotionTool promotionTool.SortPromotionTool = promotionToolList
	sort.Sort(&sortPromotionTool)

	//区分活动
	singleCampaignList := make([]*promotionTool.PromotionTool, 0, len(promotionToolIdList))
	for i := range sortPromotionTool {
		switch sortPromotionTool[i].PromotionalTypeCategory {
		case "single":
			singleCampaignList = append(singleCampaignList, sortPromotionTool[i])
		}
	}

	if len(singleCampaignList) > 0 {
		m.singleCampaignCompute(singleCampaignList)
	}
}

func (m *CampaignService) singleCampaignCompute(singleCampaignList []*promotionTool.PromotionTool) {
	processedProductIdList := make([]uint64, 0, 16)
	for i := range singleCampaignList {
		switch singleCampaignList[i].PromotionalTypeId {
		case 1:
			m.limitedTimeDown(singleCampaignList[i], &processedProductIdList)
		}
	}
}

func (m *CampaignService) limitedTimeDown(promotionToolInfo *promotionTool.PromotionTool, processedProductIdList *[]uint64) {
	campaignProductList := m.CampaignProductMap[promotionToolInfo.PromotionalId]
	for i := range campaignProductList {
		for k := range m.RequestParam.ProductList {
			if tools.InUint64(m.RequestParam.ProductList[k].ProductId, *processedProductIdList) {
				continue
			}
			if campaignProductList[i].ProductId != m.RequestParam.ProductList[k].ProductId {
				continue
			}
			*processedProductIdList = append(*processedProductIdList, campaignProductList[i].ProductId)

			var singleDiscountDecimal decimal.Decimal
			singleDiscountDecimal = decimal.NewFromFloat(m.CartProductInfoMap[campaignProductList[i].ProductId].SalePrice).Sub(decimal.NewFromFloat(campaignProductList[i].PromotionPrice))
			discount := singleDiscountDecimal.Mul(decimal.NewFromFloat(float64(m.RequestParam.ProductList[k].Quantity)))
			m.CartProductInfoMap[campaignProductList[i].ProductId].SalePrice = campaignProductList[i].PromotionPrice
			m.setResponseProductDiscount(campaignProductList[i].ProductId, discount, campaignProductList[i].PromotionPrice, promotionToolInfo)
		}
	}
}

func (m *CampaignService) setResponsePromotionList(promotionToolList []*promotionTool.PromotionTool) {
	if len(promotionToolList) == 0 {
		return
	}
	for i := range promotionToolList {
		responsePromotionList := promotionTool.ResponsePromotionList{}
		responsePromotionList.PromotionalId = promotionToolList[i].PromotionalId
		responsePromotionList.PromotionalName = promotionToolList[i].PromotionalName
		responsePromotionList.MatchStatus = 2
		responsePromotionList.MatchProducts = make([]*promotionTool.ResponseMatchProduct, 0, 8)
		m.PromotionList = append(m.PromotionList, &responsePromotionList)
	}
}

func (m *CampaignService) setResponseProductDiscount(productId uint64, discountDecimal decimal.Decimal, cartPrice float64, promotionInfo *promotionTool.PromotionTool) {
	discountFloat, _ := discountDecimal.Float64()
	responseProductDiscount := product.ResponseProductDiscount{}
	if tools.InUint64(productId, m.ProductDiscountIdList) {
		for i := range m.ProductDiscount {
			if m.ProductDiscount[i].ProductId != productId {
				continue
			}
			responseProductDiscount.CartPrice = cartPrice
			responseProductDiscount.TotalDiscount, _ = decimal.NewFromFloat(responseProductDiscount.TotalDiscount).Add(discountDecimal).Float64()
			responseProductDiscount.Promotions = append(responseProductDiscount.Promotions, &promotionTool.ResponsePromotionDiscount{
				PromotionalId:   promotionInfo.PromotionalId,
				PromotionalName: promotionInfo.PromotionalName,
				Discount:        discountFloat,
			})
		}
	} else {
		responsePromotionDiscount := make([]*promotionTool.ResponsePromotionDiscount, 0, 16)
		responseProductDiscount.ProductId = productId
		responseProductDiscount.CartPrice = cartPrice
		responseProductDiscount.TotalDiscount = discountFloat
		responsePromotionDiscount = append(responsePromotionDiscount, &promotionTool.ResponsePromotionDiscount{
			PromotionalId:   promotionInfo.PromotionalId,
			PromotionalName: promotionInfo.PromotionalName,
			Discount:        discountFloat,
		})
		responseProductDiscount.Promotions = responsePromotionDiscount
		m.ProductDiscountIdList = append(m.ProductDiscountIdList, productId)
	}
	m.ProductDiscount = append(m.ProductDiscount, &responseProductDiscount)

	for i := range m.PromotionList {
		if m.PromotionList[i].PromotionalId != promotionInfo.PromotionalId {
			continue
		}
		m.PromotionList[i].MatchStatus = 1
		m.PromotionList[i].Discount, _ = decimal.NewFromFloat(m.PromotionList[i].Discount).Add(discountDecimal).Float64()
		m.PromotionList[i].MatchProducts = append(m.PromotionList[i].MatchProducts, &promotionTool.ResponseMatchProduct{
			ProductId: productId,
			Discount:  discountFloat,
		})
	}
	m.TotalDiscount, _ = decimal.NewFromFloat(m.TotalDiscount).Add(discountDecimal).Float64()
	m.PromotionDiscount, _ = decimal.NewFromFloat(m.PromotionDiscount).Add(discountDecimal).Float64()
}

func (m *CampaignService) buildUpResponse() (result *ResponseDiscountList) {
	shouldPayment, _ := decimal.NewFromFloat(m.TotalPrice).Sub(decimal.NewFromFloat(m.TotalDiscount)).Float64()
	freight := m.RequestParam.Freight
	if shouldPayment > m.RequestParam.FreightCost {
		freight = 0
	}

	return &ResponseDiscountList{
		TotalPrice:        m.TotalPrice,
		TotalDiscount:     m.TotalDiscount,
		PromotionDiscount: m.PromotionDiscount,
		CouponDiscount:    0,
		ProductDiscount:   m.ProductDiscount,
		PromotionList:     m.PromotionList,
		Freight:           freight,
		ShouldPayment:     shouldPayment,
	}
}
