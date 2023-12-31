using UnityEngine;
using System.Collections;

/// <summary>
/// Test init class for this proj, setup some demo builiding and graphic card
/// </summary>
public class TestInit : MonoBehaviour
{
	
	// Use this for initialization
	void Start()
	{
		
		Player player = Player.Instance;
		player.Money = 50000;
        BuildingManager buildingManager = BuildingManager.Instance;
        Building demoRoom = buildingManager.FindBuildingByName("DemoRoom");

		ItemManager itemManager = ItemManager.Instance;
		GraphicCardItem card =  itemManager.FindGraphicCardItemByName("GTX1060");
		for(int i = 0; i < 1; i++)
		{
			demoRoom.AddingGraphicCard(card);
			Debug.Log(card  + " and " + demoRoom + " correct");
		}

		player.currBuildingAt = demoRoom;
		player.Buildings.Add(demoRoom);
	}

	// Update is called once per frame
	void Update()
	{
			
	}
}

