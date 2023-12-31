using UnityEngine;
using System.Collections;
using UnityEngine.UI;
using TMPro;
using static UnityEditor.Progress;

public class StoreItemSlot : MonoBehaviour
{
	GraphicCardItem Item;
	public Image Icon;
	public TextMeshProUGUI MoneyText;
	public Button Button;
	Player _player;
	public void AddItem(GraphicCardItem item)
	{
		_player = Player.Instance;
		this.Item = item	;
		this.Icon.sprite = this.Item.Icon;
		this.Icon.enabled = true;
        MoneyText.text = "$ "+Item.Price + "";
        MoneyText.enabled = true;

		//when click the item, bought it!
		Button.onClick.AddListener(OnBuy);

    }
	public void ClearSlot()
	{
		Item = null;
		Icon.enabled = false;
		Icon.sprite = null;
        MoneyText.enabled = false;
        MoneyText.text = "";

    }
    // Use this for initialization
    void Start()
	{

	}

	// Update is called once per frame
	void Update()
	{
			
	}
	void OnBuy()
	{
		if(_player.Money < Item.Price)
        {
			//TODO no money you can't buy
			Debug.Log("no money");
			return;
		}
		Building building = _player.currBuildingAt;

		Debug.Log(building.Capacity);
        Debug.Log(building.CardSize());

        if (building.Capacity < building.CardSize())
		{
            //todo no mroe capacity
            Debug.Log("no capacity");
            return;
		}
		_player.currBuildingAt.AddingGraphicCard(this.Item);
		_player.Money -= this.Item.Price;
		//TODO animation loading decreasing value smoothing papaap
	}

}

