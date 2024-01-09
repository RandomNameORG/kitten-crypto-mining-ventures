using UnityEngine;
using System.Collections;
using System.Collections.Generic;
using UnityEngine.SceneManagement;
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
	public Dictionary<Object, int> items = new Dictionary<Object, int>();

	private float passedTime; // default 0
    public float targetTime = 5.0f;  // set time interval
	

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
    void Update()
	{
		if(passedTime>targetTime)
        {
            if(items.Count != 0){
				finishbuy();
				items.Clear();
			}
            
            passedTime = 0;
        }
        passedTime += Time.deltaTime;
	}
	void OnDisable(){
		if(items.Count != 0){
			finishbuy();
			items.Clear();
		}
	}
	public void finishbuy(){
		new package(items);
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

        
		
		_player.currBuildingAt.AddingGraphicCard(this.Item);
		_player.Money -= this.Item.Price;
		if (items.ContainsKey(Item))
		{
			items[Item] += 1;
		}
		else
		{
			items[Item] = 1;
		}
		//TODO animation loading decreasing value smoothing papaap
	}

}

