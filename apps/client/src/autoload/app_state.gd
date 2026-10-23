extends Node
## App 全局状态（自动加载单例 App）：会话、服务定位、屏幕注册表与导航门面。

const SCREENS := {
	"splash": "res://src/ui/screens/splash/splash.tscn",
	"login": "res://src/ui/screens/login/login.tscn",
	"shell": "res://src/ui/screens/shell/shell.tscn",
	"home": "res://src/ui/screens/home/home.tscn",
	"errand": "res://src/ui/screens/errand/errand.tscn",
	"orders": "res://src/ui/screens/orders/orders.tscn",
	"profile": "res://src/ui/screens/profile/profile.tscn",
	"search": "res://src/ui/screens/search/search.tscn",
	"merchant": "res://src/ui/screens/merchant/merchant.tscn",
	"order_confirm": "res://src/ui/screens/order_confirm/order_confirm.tscn",
	"pay_success": "res://src/ui/screens/pay_success/pay_success.tscn",
	"order_track": "res://src/ui/screens/order_track/order_track.tscn",
	"review": "res://src/ui/screens/review/review.tscn",
	"address": "res://src/ui/screens/address/address.tscn",
	"address_edit": "res://src/ui/screens/address_edit/address_edit.tscn",
	"coupons": "res://src/ui/screens/coupons/coupons.tscn",
}

var session: SessionModel
var auth: AuthService
var demo: DemoData
var navigator: Navigator

func _ready() -> void:
	session = SessionModel.new()
	session.load()
	auth = AuthService.new()
	demo = DemoData.new()

## 打开屏幕：加载场景、构造对应 ViewModel 并压栈。
func open_screen(name: String, vm: ViewModel = null) -> ScreenBase:
	var scn: PackedScene = load(SCREENS[name])
	if scn == null:
		push_error("App.open_screen: 场景不存在 " + name)
		return null
	var target := vm if vm != null else create_vm(name)
	return navigator.push(scn, target)

## 清空导航栈并打开屏幕（登录成功进入主界面、登出回登录页）。
func replace_screen(name: String, vm: ViewModel = null) -> ScreenBase:
	var scn: PackedScene = load(SCREENS[name])
	if scn == null:
		push_error("App.replace_screen: 场景不存在 " + name)
		return null
	var target := vm if vm != null else create_vm(name)
	return navigator.replace(scn, target)

func go_back() -> void:
	navigator.pop()

## 回到主界面并切换到指定 Tab（0首页 1跑腿 2订单 3我的）。
func go_shell_tab(index: int) -> void:
	replace_screen("shell")
	var scr: ScreenBase = navigator.current()
	if scr and scr is ShellView:
		scr.switch_to(index)

func create_vm(name: String) -> ViewModel:
	match name:
		"splash":
			return SplashVM.new()
		"login":
			return LoginVM.new()
		"shell":
			return ShellVM.new()
		"home":
			return HomeVM.new()
		"errand":
			return ErrandVM.new()
		"orders":
			return OrdersVM.new()
		"profile":
			return ProfileVM.new()
		"search":
			return SearchVM.new()
		"merchant":
			return MerchantVM.new()
		"order_confirm":
			return OrderConfirmVM.new()
		"pay_success":
			return PaySuccessVM.new()
		"order_track":
			return OrderTrackVM.new()
		"review":
			return ReviewVM.new()
		"address":
			return AddressVM.new()
		"address_edit":
			return AddressEditVM.new()
		"coupons":
			return CouponsVM.new()
		_:
			push_error("App.create_vm: 未知屏幕 " + name)
			return ViewModel.new()
