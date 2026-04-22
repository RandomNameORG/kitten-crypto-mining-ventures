extends Control

const StoreItemSlot := preload("res://scenes/ui/store_item_slot.tscn")

@onready var _grid: GridContainer = $Layout/Scroll/Grid
@onready var _money_label: Label = $Layout/Header/MoneyLabel
@onready var _back_button: Button = $Layout/Header/BackButton

var _slots: Array = []

func _ready() -> void:
	_back_button.pressed.connect(MenuManager.open_main_menu)
	PlayerManager.money_changed.connect(_on_money_changed)
	_populate()
	_refresh_money()

func _populate() -> void:
	for child in _grid.get_children():
		child.queue_free()
	_slots.clear()
	for card in GraphicCardManager.get_cards():
		var slot = StoreItemSlot.instantiate()
		_grid.add_child(slot)
		slot.set_card(card)
		slot.purchase_requested.connect(_on_purchase_requested)
		_slots.append(slot)

func _on_purchase_requested(card_id: String) -> void:
	var card := GraphicCardManager.find_by_id(card_id)
	if card == null or DataManager.player == null:
		return
	if DataManager.player.money < card.price:
		return
	DataManager.player.money -= card.price
	GraphicCardManager.increment_quantity(card_id)
	_refresh_money()
	for slot in _slots:
		slot.refresh()

func _on_money_changed(_amount: int) -> void:
	_refresh_money()

func _refresh_money() -> void:
	var money: int = 0
	if DataManager.player != null:
		money = DataManager.player.money
	_money_label.text = "Money: $%d" % money
