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
	"strconv"
	"sync"

	"github.com/jinzhu/gorm"

	"github.com/shopspring/decimal"
)

var LimitedDown map[string]*sync.Map

type CampaignService struct {
	RequestParam          *promotionTool.RequestPromotionParam
	CampaignProductMap    map[uint64][]*promotionProduct.PromotionProduct
	CartProductInfoMap    map[uint64]*product.PromotionProductInfo
	ResponseDiscountList  *ResponseDiscountList
	TotalDiscount         float64
	PromotionDiscount     float64
	ProductDiscountIdList []uint64
	LimitedDownList       map[uint64]*promotionProduct.PromotionStock
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
		LimitedDownList:       make(map[uint64]*promotionProduct.PromotionStock, 8),
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
	promotionToolIdListLen := len(promotionToolIdList)
	if promotionToolIdListLen == 0 {
		return
	}

	//对有效的活动排序
	var sortPromotionTool promotionTool.SortPromotionTool = promotionToolList
	sort.Sort(&sortPromotionTool)

	//区分活动
	singleCampaignList := make([]*promotionTool.PromotionTool, 0, promotionToolIdListLen)
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
	if promotionInfo.PromotionalTypeId == 1 {
		var success uint8
		key := strconv.FormatUint(productId, 10) + ":" + strconv.FormatUint(promotionInfo.PromotionalId, 10)
		v, ok := LimitedDown[key]
		if !ok {
			goto SKIP
		}
		stockValueInter, ok := v.Load("stock")
		if !ok {
			goto SKIP
		}
		stockInfo, ok := stockValueInter.(*promotionProduct.PromotionStock)
		if !ok {
			goto SKIP
		}
		if len(m.RequestParam.OperatingType) > 0 && stockInfo.StockLimit == 2 && stockInfo.Stock > 0 {
			nowStock := stockInfo.Stock - int64(m.CartProductInfoMap[productId].Quantity)
			if nowStock >= 0 {
				success = 1
				stockInfo.Stock = nowStock
				v.Store("stock", stockInfo)
			}
		}

		m.LimitedDownList[productId] = &promotionProduct.PromotionStock{
			Stock:      stockInfo.Stock,
			StockLimit: stockInfo.StockLimit,
			ProductId:  productId,
			Success:    success,
		}
	}

SKIP:
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
		LimitedDownList:   m.LimitedDownList,
		ProductDiscount:   m.ProductDiscount,
		PromotionList:     m.PromotionList,
		Freight:           freight,
		ShouldPayment:     shouldPayment,
	}
}

func GetProductPromotionPrice(result map[uint64]float64, memberId uint64, isNewMember uint8) (map[uint64]float64, error) {
	//获取有效活动 并 排序
	promotionToolList, promotionIdList, err := GetValidCampaign(1, memberId, isNewMember)
	if err != nil {
		return result, err
	}

	//验证会员等级
	err = GetValidMemberLevel(&promotionToolList, &promotionIdList, memberId)
	if err != nil {
		return result, err
	}

	//验证平台
	err = GetValidPlatform(&promotionToolList, &promotionIdList, "app")
	if err != nil {
		return result, err
	}
	if len(promotionIdList) == 0 {
		return result, nil
	}

	var sortPromotionTool promotionTool.SortPromotionTool = promotionToolList
	sort.Sort(&sortPromotionTool)

	//获取有效活动 下商品
	promotionProductList := make([]*promotionProduct.PromotionProduct, 0, 16)
	err = connection.Db.Table(promotionProduct.GetTableName()).Select(promotionProduct.GetField()).
		Where("promotional_id in (?) AND is_add = ?", promotionIdList, 1).Find(&promotionProductList).Error
	if err != nil {
		return result, err
	}
	promotionProductListMap := make(map[uint64][]*promotionProduct.PromotionProduct, len(promotionProductList))
	for k := range promotionProductList {
		promotionProductListMap[promotionProductList[k].PromotionalId] = append(promotionProductListMap[promotionProductList[k].PromotionalId], promotionProductList[k])
	}

	for k := range sortPromotionTool {
		//遍历 活动下商品
		for i := range promotionProductListMap[sortPromotionTool[k].PromotionalId] {
			v, ok := result[promotionProductListMap[sortPromotionTool[k].PromotionalId][i].ProductId]
			if !ok || v > 0 {
				continue
			}

			result[promotionProductList[k].ProductId] = promotionProductList[k].PromotionPrice
		}
	}

	return result, nil
}

func InitPromotionStock() error {
	//获取结束时间大于当前时间(也就是有效的)限时直降活动
	nowTimeUnix := tools.GetNow().Unix()
	promotionToolIdList := make([]uint64, 0, 32)
	err := connection.Db.Table(promotionTool.GetTableName()).
		Where("promotional_type_id = ? AND end_time > ? AND marketing_platform = ? AND business_id = ? AND status = ?", 1, nowTimeUnix, 1, 1, 2).
		Pluck("promotional_id", &promotionToolIdList).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return err
	}
	promotionToolIdListLen := len(promotionToolIdList)
	if promotionToolIdListLen == 0 {
		return nil
	}

	//查询这个活动下的商品
	promotionProductList := make([]*promotionProduct.PromotionProduct, 0, 32)
	err = connection.Db.Table(promotionProduct.GetTableName()).Select(promotionProduct.GetField()).
		Where("promotional_id in (?) AND is_add = ?", promotionToolIdList, 1).Find(&promotionProductList).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return nil
	}
	promotionProductListLen := len(promotionProductList)

	LimitedDown = make(map[string]*sync.Map, promotionProductListLen)
	for k := range promotionProductList {
		key := strconv.FormatUint(promotionProductList[k].ProductId, 10) + ":" + strconv.FormatUint(promotionProductList[k].PromotionalId, 10)

		newSyncMap := new(sync.Map)
		newSyncMap.Store("stock", &promotionProduct.PromotionStock{
			StockLimit: promotionProductList[k].StockLimit,
			Stock:      int64(promotionProductList[k].Stock),
			ProductId:  promotionProductList[k].ProductId,
		})

		LimitedDown[key] = newSyncMap
	}

	return nil
}
