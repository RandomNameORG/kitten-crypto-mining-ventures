extends Node

signal building_changed(id: String)
signal cards_changed(building_id: String)

var placed_cards: Dictionary = {}

func find_by_id(id: String) -> RefCounted:
	for b in DataManager.buildings:
		if b.id == id:
			return b
	return null

func find_by_name(name: String) -> RefCounted:
	for b in DataManager.buildings:
		if b.name == name:
			return b
	return null

var current_building: RefCounted:
	get:
		if DataManager.player == null:
			if DataManager.buildings.size() > 0:
				return DataManager.buildings[0]
			return null
		var cid: String = DataManager.player.curr_building_id
		var found := find_by_id(cid)
		if found != null:
			return found
		if DataManager.buildings.size() > 0:
			return DataManager.buildings[0]
		return null

func set_current_building(id: String) -> void:
	if DataManager.player == null:
		return
	var b := find_by_id(id)
	if b == null:
		return
	DataManager.player.curr_building_id = b.id
	DataManager.player.curr_building_name = b.name
	building_changed.emit(b.id)

func _cards_for(building_id: String) -> Array:
	if not placed_cards.has(building_id):
		return []
	return placed_cards[building_id]

func place_card(building_id: String, card_id: String, grid_pos: Vector2i) -> void:
	if not placed_cards.has(building_id):
		placed_cards[building_id] = []
	placed_cards[building_id].append({"id": card_id, "pos": grid_pos})
	cards_changed.emit(building_id)

func remove_card(building_id: String, index: int) -> void:
	if not placed_cards.has(building_id):
		return
	var arr: Array = placed_cards[building_id]
	if index < 0 or index >= arr.size():
		return
	arr.remove_at(index)
	cards_changed.emit(building_id)

func get_money_per_second() -> int:
	var b := current_building
	if b == null:
		return 0
	var cards: Array = _cards_for(b.id)
	if cards.is_empty():
		return b.money_per_second
	var total: int = 0
	for entry in cards:
		var cid: String = entry.get("id", "") if typeof(entry) == TYPE_DICTIONARY else ""
		var card := GraphicCardManager.find_by_id(cid)
		if card != null:
			total += card.per_second_earn
	return total

func get_volt_per_second() -> int:
	var b := current_building
	if b == null:
		return 0
	var cards: Array = _cards_for(b.id)
	if cards.is_empty():
		return b.volt_per_second
	var total: int = 0
	for entry in cards:
		var cid: String = entry.get("id", "") if typeof(entry) == TYPE_DICTIONARY else ""
		var card := GraphicCardManager.find_by_id(cid)
		if card != null:
			total += card.per_second_lose_volt
	return total
