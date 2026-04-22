extends Node

signal money_changed(amount: int)

var timer: float = 0.0
var display_money: float = 0.0
var _tween: Tween

func _process(delta: float) -> void:
	timer += delta
	if timer >= 1.0:
		_per_second_earn_money()
		timer -= 1.0

func _per_second_earn_money() -> void:
	if DataManager.player == null:
		return
	var total: int = 0
	for bref in DataManager.player.buildings:
		var ref_id = bref.get("id", "") if typeof(bref) == TYPE_DICTIONARY else ""
		for b in DataManager.buildings:
			if b.id == ref_id:
				total += b.money_per_second
				break
	var pre: int = DataManager.player.money
	DataManager.player.money += total
	if _tween and _tween.is_valid():
		_tween.kill()
	_tween = create_tween()
	_tween.tween_method(_update_display_money, float(pre), float(DataManager.player.money), 0.1)

func _update_display_money(v: float) -> void:
	display_money = floor(v)
	money_changed.emit(int(display_money))

func _current_building() -> RefCounted:
	if DataManager.player == null:
		return null
	var cid := DataManager.player.curr_building_id
	for b in DataManager.buildings:
		if b.id == cid:
			return b
	if DataManager.buildings.size() > 0:
		return DataManager.buildings[0]
	return null

func get_current_volt() -> int:
	var b := _current_building()
	if b == null:
		return 0
	return b.volt_per_second

func get_max_volt() -> int:
	var b := _current_building()
	if b == null:
		return 0
	return b.max_volt
