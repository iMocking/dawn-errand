class_name PaySuccessVM
extends ViewModel
## 支付成功 VM：订单号、送达时间、地址；查看订单/返回首页。

var _order_no := ""
var _shop := ""
var _total := 0.0
var _address: AddressModel

func _init(shop: String = "", order_no: String = "", total: float = 0.0, addr: AddressModel = null) -> void:
	_shop = shop
	_order_no = order_no
	_total = total
	_address = addr
	obs("shop", shop)
	obs("order_no", order_no)
	obs("total", Fmt.yuan(total))
	obs("eta", "预计 12:45 送达")
	obs("address_name", addr.name if addr else "王建国")
	obs("address_detail", (addr.province + addr.detail) if addr else "王家湾村三组26号")

func go_orders() -> void:
	App.go_shell_tab(2)

func go_home() -> void:
	App.go_shell_tab(0)
