class_name AddressModel
extends RefCounted
## 收货地址模型（地址管理）。

var id: String = ""
var name: String = ""
var phone: String = ""
var province: String = "湖北省黄冈市浠水县"
var detail: String = ""
var landmark: String = ""
var tag: String = "家"
var is_default: bool = false

func full_detail() -> String:
	var s := province + "\n" + detail
	if landmark != "":
		s += " · " + landmark
	return s
