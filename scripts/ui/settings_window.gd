extends Control

const RESOLUTIONS: Array = [
	Vector2i(1280, 720),
	Vector2i(1600, 900),
	Vector2i(1920, 1080),
	Vector2i(2560, 1440),
]

@onready var _resolution: OptionButton = $Panel/Layout/Resolution/OptionButton
@onready var _master_slider: HSlider = $Panel/Layout/Master/Slider
@onready var _sfx_slider: HSlider = $Panel/Layout/Sfx/Slider
@onready var _back_button: Button = $Panel/Layout/BackButton

func _ready() -> void:
	_populate_resolutions()
	_master_slider.value = _bus_linear("Master")
	_sfx_slider.value = _bus_linear("SFX")
	_master_slider.value_changed.connect(_on_master_changed)
	_sfx_slider.value_changed.connect(_on_sfx_changed)
	_resolution.item_selected.connect(_on_resolution_selected)
	_back_button.pressed.connect(MenuManager.open_main_menu)

func _populate_resolutions() -> void:
	_resolution.clear()
	var current := DisplayServer.window_get_size()
	var select_idx := 0
	for i in RESOLUTIONS.size():
		var r: Vector2i = RESOLUTIONS[i]
		_resolution.add_item("%d x %d" % [r.x, r.y], i)
		if r == current:
			select_idx = i
	_resolution.select(select_idx)

func _bus_linear(name: String) -> float:
	var idx := AudioServer.get_bus_index(name)
	if idx < 0:
		return 1.0
	return db_to_linear(AudioServer.get_bus_volume_db(idx))

func _set_bus_linear(name: String, value: float) -> void:
	var idx := AudioServer.get_bus_index(name)
	if idx < 0:
		return
	AudioServer.set_bus_volume_db(idx, linear_to_db(maxf(value, 0.0001)))

func _on_master_changed(value: float) -> void:
	_set_bus_linear("Master", value)

func _on_sfx_changed(value: float) -> void:
	_set_bus_linear("SFX", value)

func _on_resolution_selected(index: int) -> void:
	if index < 0 or index >= RESOLUTIONS.size():
		return
	DisplayServer.window_set_size(RESOLUTIONS[index])
