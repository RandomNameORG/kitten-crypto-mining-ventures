extends Node

const GraphicCardModel = preload("res://scripts/models/graphic_card.gd")

var cards: Array = []
var _quantities: Dictionary = {}

func _ready() -> void:
	cards = DataManager.graphic_cards.duplicate()
	for c in cards:
		_quantities[c.id] = c.quantity

func find_by_id(id: String) -> RefCounted:
	for c in cards:
		if c.id == id:
			return c
	return null

func find_by_name(name: String) -> RefCounted:
	for c in cards:
		if c.name == name:
			return c
	return null

func get_cards() -> Array:
	return cards

func instantiate_card(id: String) -> RefCounted:
	var src := find_by_id(id)
	if src == null:
		return null
	var copy = GraphicCardModel.new()
	copy.from_dict(src.to_dict())
	return copy

func get_quantity(id: String) -> int:
	return int(_quantities.get(id, 0))

func increment_quantity(id: String) -> void:
	_quantities[id] = get_quantity(id) + 1
	var c := find_by_id(id)
	if c != null:
		c.quantity = _quantities[id]

func decrement_quantity(id: String) -> void:
	var next := get_quantity(id) - 1
	if next < 0:
		next = 0
	_quantities[id] = next
	var c := find_by_id(id)
	if c != null:
		c.quantity = _quantities[id]
