class_name ShellVM
extends ViewModel
## 主界面外壳 VM：当前 Tab（首页/跑腿/订单/我的）。

func _init() -> void:
	obs("tab", 0)

func switch_tab(index: int) -> void:
	set_prop("tab", index)
