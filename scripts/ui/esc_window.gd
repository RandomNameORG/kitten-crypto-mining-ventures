extends CanvasLayer

@onready var _root: Control = $Root
@onready var _resume_button: Button = $Root/Center/Panel/Buttons/ResumeButton
@onready var _settings_button: Button = $Root/Center/Panel/Buttons/SettingsButton
@onready var _quit_button: Button = $Root/Center/Panel/Buttons/QuitButton

func _ready() -> void:
	process_mode = Node.PROCESS_MODE_ALWAYS
	_root.visible = false
	_resume_button.pressed.connect(_on_resume)
	_settings_button.pressed.connect(MenuManager.open_settings)
	_quit_button.pressed.connect(MenuManager.open_main_menu)

func _unhandled_input(event: InputEvent) -> void:
	if event is InputEventKey and event.pressed and not event.echo and event.keycode == KEY_ESCAPE:
		_toggle()
		get_viewport().set_input_as_handled()

func _toggle() -> void:
	var next := not _root.visible
	_root.visible = next
	MenuManager.toggle_esc(next)

func _on_resume() -> void:
	_root.visible = false
	MenuManager.toggle_esc(false)
