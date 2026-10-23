class_name Command
extends RefCounted
## MVVM 核心：命令封装。将「可执行逻辑 + 执行中状态」暴露给 View，
## 支持防重复提交（is_running 由 execute 回调返回的异步状态管理）。

signal executed
signal failed(message: String)

var _fn: Callable
var can_execute: bool = true

func _init(fn: Callable) -> void:
	_fn = fn

func execute() -> void:
	if not can_execute:
		return
	if _fn.is_valid():
		_fn.call()
