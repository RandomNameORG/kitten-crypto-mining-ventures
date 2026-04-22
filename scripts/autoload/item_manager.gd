extends Node

var _items: Dictionary = {}

func add(id: String, qty: int = 1) -> void:
	_items[id] = count(id) + qty

func remove(id: String, qty: int = 1) -> void:
	var next := count(id) - qty
	if next <= 0:
		_items.erase(id)
	else:
		_items[id] = next

func has(id: String) -> bool:
	return count(id) > 0

func count(id: String) -> int:
	return int(_items.get(id, 0))
