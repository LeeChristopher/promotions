package filters

import (
	"encoding/json"
	"errors"
	"net/http"
	"promotions/models/promotionProduct"
	"promotions/models/promotionTool"
	"promotions/services"
	"regexp"
	"strconv"

	"github.com/astaxie/beego/validation"
)

type CampaignFilter struct {
	Request *http.Request
}

func NewCampaignFilter(request *http.Request) *CampaignFilter {
	return &CampaignFilter{Request: request}
}

func (m *CampaignFilter) GetDiscountList() (list []string, err error) {
	businessKey := m.Request.FormValue("business_key")
	memberId := m.Request.FormValue("member_id")
	platform := m.Request.FormValue("platform")
	productList := m.Request.FormValue("product_list")
	isNewMember := m.Request.FormValue("is_new_member")
	freight := m.Request.FormValue("freight")
	freightCost := m.Request.FormValue("freight_cost")

	valid := validation.Validation{}
	valid.Required(businessKey, "business_key").Message("请提交商户信息")
	valid.Match(businessKey, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`), "business_key").Message("商户信息格式错误！")
	valid.Required(memberId, "member_id").Message("请提交用户信息")
	valid.Match(memberId, regexp.MustCompile(`^[0-9]+$`), "member_id").Message("用户信息格式错误！")
	valid.Required(platform, "platform").Message("请提交平台信息")
	valid.Match(platform, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`), "platform").Message("平台信息格式错误！")
	valid.Required(productList, "product_list").Message("请提交商品信息")
	valid.Required(isNewMember, "is_new_member").Message("请提交用户新旧信息！")
	valid.Match(isNewMember, regexp.MustCompile(`^1|2$`), "is_new_member").Message("用户新旧信息格式错误！")
	valid.Required(freight, "freight").Message("请提交运费信息！")
	valid.Match(freight, regexp.MustCompile(`^[0-9]$`), "freight").Message("运费信息格式错误！")
	valid.Required(freightCost, "freight_cost").Message("请提交运费门槛信息！")
	valid.Match(freightCost, regexp.MustCompile(`^[1-9][0-9]*$`), "freight_cost").Message("运费门槛信息格式错误！")
	if valid.HasErrors() {
		return nil, valid.Errors[0]
	}

	cartProductList := make([]*promotionProduct.RequestPromotionProduct, 0, 32)
	err = json.Unmarshal([]byte(productList), &cartProductList)
	if err != nil {
		return nil, errors.New("商品信息格式错误！")
	}
	memberIdInt, err := strconv.Atoi(memberId)
	if err != nil {
		return nil, errors.New("用户信息格式错误！")
	}
	isNewMemberInt, err := strconv.Atoi(isNewMember)
	if err != nil {
		return nil, errors.New("用户新旧信息格式错误！")
	}
	freightInt, err := strconv.Atoi(freight)
	if err != nil {
		return nil, errors.New("运费信息格式错误！")
	}
	freightCostInt, err := strconv.Atoi(freightCost)
	if err != nil {
		return nil, errors.New("运费门槛信息格式错误！")
	}
	memberUint := uint64(memberIdInt)
	isNewMemberUint := uint8(isNewMemberInt)
	freightUint := uint64(freightInt)
	freightCostUint := uint64(freightCostInt)

	requestPromotionParam := &promotionTool.RequestPromotionParam{
		BusinessKey: businessKey,
		MemberId:    memberUint,
		Platform:    platform,
		ProductList: cartProductList,
		IsNewMember: isNewMemberUint,
		Freight:     freightUint,
		FreightCost: freightCostUint,
	}
	campaignService := services.NewCampaign(requestPromotionParam)
	list, err = campaignService.GetDiscountList()

	return list, nil
}
