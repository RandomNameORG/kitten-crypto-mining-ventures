extends PanelContainer

signal purchase_requested(card_id: String)

@onready var _icon: TextureRect = $Layout/Icon
@onready var _name_label: Label = $Layout/Info/NameLabel
@onready var _price_label: Label = $Layout/Info/PriceLabel
@onready var _earn_label: Label = $Layout/Info/EarnLabel
@onready var _volt_label: Label = $Layout/Info/VoltLabel
@onready var _quantity_label: Label = $Layout/Info/QuantityLabel
@onready var _buy_button: Button = $Layout/Info/BuyButton

var _card_id: String = ""

func _ready() -> void:
	_buy_button.pressed.connect(_on_buy_pressed)

func set_card(card: GraphicCard) -> void:
	_card_id = card.id
	_name_label.text = card.name
	_price_label.text = "Price: $%d" % card.price
	_earn_label.text = "Earn: $%d/s" % card.per_second_earn
	_volt_label.text = "Volt: %d/s" % card.per_second_lose_volt
	_update_quantity(card)
	if card.image_source_path != "":
		var res_path := "res://assets/art/cards/%s.png" % card.image_source_path
		if ResourceLoader.exists(res_path):
			var tex := load(res_path)
			if tex != null:
				_icon.texture = tex

func refresh() -> void:
	var card := GraphicCardManager.find_by_id(_card_id)
	if card != null:
		_update_quantity(card)

func _update_quantity(card: GraphicCard) -> void:
	_quantity_label.text = "Owned: %d" % GraphicCardManager.get_quantity(card.id)
	_buy_button.disabled = card.is_locked

func _on_buy_pressed() -> void:
	purchase_requested.emit(_card_id)
