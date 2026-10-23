class_name ErrandModel
extends RefCounted
## 跑腿订单模型（跑腿下单页）。

var type: String = "帮我买"     # 帮我买/帮我送/帮我取/帮我办
var from_label: String = "取货地址"
var from_address: String = ""
var from_contact: String = ""
var to_label: String = "收货地址"
var to_address: String = ""
var to_contact: String = ""
var goods_desc: String = ""
var tags: Array = []
var distance_km: float = 5.2
var fee: float = 8.0
var advance_total: float = 0.0   # 商品垫付（实报实销）
