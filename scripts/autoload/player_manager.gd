extends Node

signal money_changed(amount: int)
signal volt_changed(text: String)

var timer: float = 0.0
var display_money: float = 0.0
var _tween: Tween

func _process(delta: float) -> void:
	timer += delta
	if timer >= 1.0:
		_per_second_earn_money()
		_per_second_volt_display()
		timer -= 1.0

func _per_second_earn_money() -> void:
	if DataManager.player == null:
		return
	var total: int = BuildingManager.get_money_per_second()
	var pre: int = DataManager.player.money
	DataManager.player.money += total
	if _tween and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	_tween.tween_method(_update_display_money, float(pre), float(DataManager.player.money), 0.1)

func _update_display_money(v: float) -> void:
	display_money = floor(v)
	money_changed.emit(int(display_money))

func _per_second_volt_display() -> void:
	volt_changed.emit(get_volt_display_text())

func get_volt_display_text() -> String:
	var volt := BuildingManager.get_volt_per_second()
	var b = BuildingManager.current_building
	var max_volt: int = b.max_volt if b != null else 0
	return "%d/%d" % [volt, max_volt]

func _current_building() -> RefCounted:
	return BuildingManager.current_building

func get_current_volt() -> int:
	return BuildingManager.get_volt_per_second()

func get_max_volt() -> int:
	var b = BuildingManager.current_building
	if b == null:
		return 0
	return b.max_volt
