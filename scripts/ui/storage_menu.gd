extends Control

const StoreItemSlot := preload("res://scenes/ui/store_item_slot.tscn")

@onready var _grid: GridContainer = $Layout/Scroll/Grid
@onready var _back_button: Button = $Layout/Header/BackButton
@onready var _empty_label: Label = $Layout/EmptyLabel

func _ready() -> void:
	_back_button.pressed.connect(_on_back_pressed)
	_populate()

func _populate() -> void:
	for child in _grid.get_children():
		child.queue_free()
	var any := false
	for card in GraphicCardManager.get_cards():
		if GraphicCardManager.get_quantity(card.id) <= 0:
			continue
		any = true
		var slot = StoreItemSlot.instantiate()
		_grid.add_child(slot)
		slot.set_card(card)
		slot.purchase_requested.connect(_on_slot_purchase)
	_empty_label.visible = not any

func _on_slot_purchase(_card_id: String) -> void:
	pass

func _on_back_pressed() -> void:
	MenuManager.resume()
	self.queue_free()
