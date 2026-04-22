extends Node

func fade_in(node: CanvasItem, duration: float) -> Tween:
	var tween := node.create_tween()
	node.modulate.a = 0.0
	tween.tween_property(node, "modulate:a", 1.0, duration)
	return tween

func fade_out(node: CanvasItem, duration: float) -> Tween:
	var tween := node.create_tween()
	tween.tween_property(node, "modulate:a", 0.0, duration)
	return tween

func fade_sequence(node: CanvasItem, fade_duration: float, wait_duration: float) -> Tween:
	var tween := node.create_tween()
	node.modulate.a = 0.0
	tween.tween_property(node, "modulate:a", 1.0, fade_duration)
	tween.tween_interval(wait_duration)
	tween.tween_property(node, "modulate:a", 0.0, fade_duration)
	return tween
