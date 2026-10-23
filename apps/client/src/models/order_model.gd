class_name OrderModel
extends RefCounted
## 订单模型（我的订单列表）。

enum StatusType { ONGOING, DONE, CANCELLED, AFTER_SALE }

var id: String = ""
var shop: String = ""
var status: String = "配送中"
var status_type: int = StatusType.ONGOING
var goods_summary: String = ""
var time_text: String = ""
var total: float = 0.0
var icon: int = 0
var actions: Array = []   # ["取消订单","查看配送"]

func status_color() -> Color:
	match status_type:
		StatusType.ONGOING:
			return Color("FF6B2C")
		StatusType.DONE:
			return Color("86909C")
		StatusType.CANCELLED:
			return Color("C9CDD4")
		StatusType.AFTER_SALE:
			return Color("FF4D4F")
	return Color("86909C")
