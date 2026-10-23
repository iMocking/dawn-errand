class_name OrderTrackVM
extends ViewModel
## 订单追踪 VM：配送状态、进度节点、骑手信息（演示数据源）。

var order: OrderModel

func _init(o: OrderModel = null) -> void:
	var src := o
	if src == null:
		var orders: Array = App.demo.orders()
		for it in orders:
			if it.status_type == OrderModel.StatusType.ONGOING:
				src = it
				break
		if src == null:
			src = orders[0]
	order = src
	obs("status", order.status)
	obs("eta", "预计 12:45 送达")
	obs("shop_name", order.shop)
	obs("rider_name", "张师傅")
	obs("rider_phone", "159****1123")
	obs("rider_plate", "鄂J·D1234")
	list("steps").set_items(["商家已接单", "骑手已取货", "配送中", "已送达"])
	obs("step_index", 2)

func call_rider() -> void:
	Toast.show_msg("拨号功能演示：" + str(get_prop("rider_phone")))

func contact_service() -> void:
	Toast.show_msg("客服热线：400-800-1234")
