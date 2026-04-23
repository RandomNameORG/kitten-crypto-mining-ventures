extends Control

const BuildingScene = preload("res://scenes/building.tscn")

@onready var money_label: Label = %MoneyLabel
@onready var volt_label: Label = %VoltLabel

func _ready() -> void:
	var building_root := BuildingScene.instantiate()
	add_child(building_root)
	move_child(building_root, 0)
	PlayerManager.money_changed.connect(_on_money_changed)
	PlayerManager.volt_changed.connect(_on_volt_changed)
	if DataManager.player != null:
		_on_money_changed(DataManager.player.money)
	_on_volt_changed(PlayerManager.get_volt_display_text())

func _on_money_changed(amount: int) -> void:
	money_label.text = "$" + str(amount)

func _on_volt_changed(text: String) -> void:
	volt_label.text = text
