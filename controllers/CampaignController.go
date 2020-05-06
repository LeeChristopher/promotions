package controllers

import (
	"promotions/filters"
	"promotions/packages/tools"
)

var campaignFilter *filters.CampaignFilter

type CampaignController struct {
	BaseController
}

func (m *CampaignController) Initialise() {
	campaignFilter = filters.NewCampaignFilter(m.Ctx.Request)
}

func (m *CampaignController) GetDiscountList() {
	list, err := campaignFilter.GetDiscountList()
	if err != nil {
		m.SetResponse(tools.CodeMap["fail"], err.Error(), nil)
		return
	}

	m.SetResponse(tools.CodeMap["success"], "", list)
}
