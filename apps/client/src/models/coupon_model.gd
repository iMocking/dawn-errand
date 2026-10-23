class_name CouponModel
extends RefCounted
## 优惠券模型（我的优惠券）。

var id: String = ""
var amount: float = 0.0
var condition: String = ""     # "满20可用" / "无门槛"
var title: String = ""
var desc: String = ""
var expire_text: String = ""
var usable: bool = true

func amount_text() -> String:
	return Fmt.money(amount)
