﻿using UnityEngine;
using System.Collections;
using System.Collections.Generic;
using UnityEngine.SceneManagement;
using UnityEngine.UI;
using TMPro;
using static UnityEditor.Progress;

public class StoreItemSlot : MonoBehaviour
{
	GraphicCard Item;
	public Image Icon;
	public TextMeshProUGUI MoneyText;
	public Button Button;
	Player _player;
	public Dictionary<object, int> Items = new Dictionary<object, int>();

	private float PassedTime; // default 0
	public float TargetTime = 5.0f;  // set time interval

	public GameObject LogPane;

	public void AddItem(GraphicCard item)
	{
		_player = Player._instance;
		this.Item = item;
		this.Icon.sprite = this.Item.Icon;
		this.Icon.enabled = true;
		MoneyText.text = "$ " + Item.Price + "";
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

		if (PassedTime > TargetTime)
		{
			if (Items.Count != 0)
			{
				finishbuy();
				Items.Clear();
			}

			PassedTime = 0;
		}
		PassedTime += Time.deltaTime;
	}
	void OnDisable()
	{
		if (Items.Count != 0)
		{
			finishbuy();
			Items.Clear();
		}
	}
	public void finishbuy()
	{
		new Package(Items);
	}


	void OnBuy()
	{
		if (_player.Money < Item.Price)
		{
			PopLogManager._instance.Show(PaneLogType.NO_ENOUGH_MONEY);
			Logger.Log("no money");
			return;
		}
		Building building = _player.CurrBuildingAt;




		_player.CurrBuildingAt.AddingGraphicCard(this.Item);
		_player.Money -= this.Item.Price;
		if (Items.ContainsKey(Item))
		{
			Items[Item] += 1;
		}
		else
		{
			Items[Item] = 1;
		}
		//TODO animation loading decreasing value smoothing papaap
	}

}

