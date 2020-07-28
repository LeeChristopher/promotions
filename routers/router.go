package routers

import (
	"promotions/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/test", &controllers.TestController{}, "get:Index")
	beego.Router("/login", &controllers.AuthController{}, "get:Login")
	beego.Router("/promotion", &controllers.CampaignController{}, "post:GetDiscountList")
	beego.Router("/get-product-price", &controllers.CampaignController{}, "post:GetProductPromotionPrice")
}
