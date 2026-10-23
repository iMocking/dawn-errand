extends Control
## 应用根节点：创建导航器并进入启动页。

func _ready() -> void:
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	App.open_screen("splash")
