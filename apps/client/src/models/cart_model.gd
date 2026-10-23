class_name CartModel
extends RefCounted
## 购物车模型（商家详情底部购物车栏 + 确认订单明细）。

var lines: Array = []   # {dish: DishModel, qty: int}
var delivery_fee: float = 3.0
var pack_fee: float = 1.5

func add(dish: DishModel, qty: int = 1) -> void:
	for line in lines:
		if line.dish == dish:
			line.qty += qty
			return
	lines.append({"dish": dish, "qty": qty})

func remove(dish: DishModel) -> void:
	for i in range(lines.size() - 1, -1, -1):
		if lines[i].dish == dish:
			lines[i].qty -= 1
			if lines[i].qty <= 0:
				lines.remove_at(i)
			return

func count() -> int:
	var n := 0
	for line in lines:
		n += int(line.qty)
	return n

func goods_total() -> float:
	var t := 0.0
	for line in lines:
		t += float(line.dish.price) * float(line.qty)
	return t

func total() -> float:
	return goods_total() + delivery_fee + pack_fee

func clear() -> void:
	lines.clear()
