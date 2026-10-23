extends Control
## 单屏验证（home 首页）：注入会话 → 首页停留 20 秒。

func _ready() -> void:
	var nav := Navigator.new()
	add_child(nav)
	App.navigator = nav
	_run()

func _run() -> void:
	await get_tree().create_timer(0.5).timeout
	App.open_screen("splash")
	await get_tree().create_timer(1.0).timeout
	App.session.apply_login({
		"token": "smoke-access-token",
		"refresh_token": "smoke-refresh-token",
		"user": {"id": "u-test-001", "name": "测试用户", "phone": "13888888888", "role": "user"},
		"permissions": [],
	})
	print("SHOT: single_home")
	await get_tree().create_timer(20.0).timeout
	get_tree().quit(0)
