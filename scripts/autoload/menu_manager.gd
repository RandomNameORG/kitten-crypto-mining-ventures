extends Node

const MAIN_MENU_PATH := "res://scenes/main_menu.tscn"
const MAIN_PATH := "res://scenes/main.tscn"
const STORE_PATH := "res://scenes/store.tscn"
const SETTINGS_PATH := "res://scenes/settings.tscn"

signal settings_requested
signal storage_requested
signal esc_toggled(visible: bool)

func change_scene(path: String) -> void:
	get_tree().change_scene_to_file(path)

func start_game() -> void:
	change_scene(MAIN_PATH)

func open_main_menu() -> void:
	resume()
	change_scene(MAIN_MENU_PATH)

func open_store() -> void:
	change_scene(STORE_PATH)

func open_settings() -> void:
	change_scene(SETTINGS_PATH)
	settings_requested.emit()

func open_storage() -> void:
	storage_requested.emit()

func quit_game() -> void:
	get_tree().quit()

func pause() -> void:
	get_tree().paused = true

func resume() -> void:
	get_tree().paused = false

func toggle_esc(is_visible: bool) -> void:
	esc_toggled.emit(is_visible)
	if is_visible:
		pause()
	else:
		resume()
