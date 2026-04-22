extends Control

@onready var money_label: Label = %MoneyLabel

func _ready() -> void:
	PlayerManager.money_changed.connect(_on_money_changed)
	if DataManager.player != null:
		_on_money_changed(DataManager.player.money)

func _on_money_changed(amount: int) -> void:
	money_label.text = "$" + str(amount)
