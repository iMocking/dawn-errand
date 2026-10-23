class_name DishModel
extends RefCounted
## 菜品模型（商家详情菜单）。

var id: String = ""
var name: String = ""
var desc: String = ""
var price: float = 0.0
var monthly_sales: int = 0
var icon: int = 0
var icon_color: Color = Color("FF8A3D")

func price_text() -> String:
	return Fmt.money(price)

func sales_text() -> String:
	return "月售" + str(monthly_sales) + "份"
