extends Control

@onready var _start_button: Button = $Center/Buttons/StartButton
@onready var _store_button: Button = $Center/Buttons/StoreButton
@onready var _settings_button: Button = $Center/Buttons/SettingsButton
@onready var _quit_button: Button = $Center/Buttons/QuitButton

func _ready() -> void:
	_store_button.pressed.connect(MenuManager.open_store)
