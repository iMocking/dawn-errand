class_name MerchantModel
extends RefCounted
## 商家模型（首页推荐/附近商家）。

var id: String = ""
var name: String = ""
var score: float = 4.8
var dist_km: float = 0.0
var monthly_sales: int = 0
var eta_min: int = 25
var tags: Array = []
var icon: int = 0          # 图标类型（对应 Icon 枚举）
var banner_color: Color
var gradient: Array = []   # 双色渐变

func score_text() -> String:
	return String.num(score, 1)

func dist_text() -> String:
	return Fmt.km(dist_km)

func meta_text() -> String:
	return "⭐ " + score_text() + " · 月售" + str(monthly_sales) + "单 · " + str(eta_min) + "分钟送到"
