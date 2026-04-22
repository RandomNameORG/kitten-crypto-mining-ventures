extends Node

func change_scene(path: String) -> void:
	get_tree().change_scene_to_file(path)

func quit_game() -> void:
	get_tree().quit()

func pause() -> void:
	get_tree().paused = true

func resume() -> void:
	get_tree().paused = false
