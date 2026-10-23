class_name SplashVM
extends ViewModel
## 启动页 VM：品牌展示后按登录态导航（已登录→主界面，否则→登录页）。

var navigating := false

func start() -> void:
	if navigating:
		return
	navigating = true
	var tree: SceneTree = Engine.get_main_loop()
	await tree.create_timer(1.4).timeout
	if App.session.is_logged_in():
		App.replace_screen("shell")
	else:
		App.replace_screen("login")
