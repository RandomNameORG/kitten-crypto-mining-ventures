using UnityEngine;
using UnityEditor;
using System.Collections.Generic;
using System.Linq;
using System;
public class DataEditor : EditorWindow
{

    [MenuItem("Tools/Game Data Editor")]
    public static void ShowWindow()
    {
        GetWindow<DataEditor>("Graphic Card Manager");
    }
    private Dictionary<DataType, object> JsonDataMap = new Dictionary<DataType, object>();
    private Dictionary<DataType, object> gameDataMap = new Dictionary<DataType, object>();

    private Tab selectedTabsIndex = Tab.Building;
    private int SelectedCardIndex = 0;

    void OnGUI()
    {
        LazyLoadData();

        var tabs = new string[] { "Building", "Graphic Cards", "Player" };
        selectedTabsIndex = (Tab)GUILayout.Toolbar((int)selectedTabsIndex, tabs);
        switch (selectedTabsIndex)
        {
            case Tab.Building:
                DrawBuildingTab();
                break;
            case Tab.Graphic_Card:
                DrawCardsTab();
                break;
            case Tab.Player:
                DrawPlayerTab();
                break;
        }
    }
    public enum Tab : int
    {
        Building,
        Graphic_Card,
        Player
    }


    /// <summary>
    /// load data when needed
    /// </summary>
    /// <exception cref="NotImplementedException"></exception>
    private void LazyLoadData()
    {
        switch (selectedTabsIndex)
        {
            case Tab.Building:
                {
                    if (JsonDataMap.ContainsKey(DataType.BuildingData)) { return; }
                    var jsonData = LoadData<BuildingEntryList>(DataType.BuildingData);
                    var temp = EditorDataMapper.BuildingJsonToData(jsonData);
                    gameDataMap.Add(DataType.BuildingData, temp.buildings);
                    break;
                }
            case Tab.Graphic_Card:
                {
                    if (JsonDataMap.ContainsKey(DataType.GraphicCardData)) { return; }
                    var jsonData = LoadData<GraphicCardList>(DataType.GraphicCardData);
                    var temp = EditorDataMapper.CardJsonToData(jsonData);
                    gameDataMap.Add(DataType.GraphicCardData, temp.cards);
                    break;
                }
            case Tab.Player:
                {
                    if (JsonDataMap.ContainsKey(DataType.PlayerData)) { return; }
                    var jsonData = LoadData<PlayerEntry>(DataType.PlayerData);
                    var temp = EditorDataMapper.PlayerJsonToData(jsonData);
                    gameDataMap.Add(DataType.PlayerData, temp);
                }
                break;
        }
    }
    private T LoadData<T>(DataType type)
    {
        var res = DataLoader.LoadData<T>(type);
        JsonDataMap.Add(type, res);
        return res;
    }

    private int SelectedBuildingIndex = 0;
    void DrawBuildingTab()
    {
        var buildings = (List<Building>)gameDataMap[DataType.BuildingData];
        SelectedBuildingIndex = EditorGUILayout.Popup(SelectedBuildingIndex, buildings.Select(e => e.Name).ToArray());
        var building = buildings[SelectedBuildingIndex];
        if (building != null)
        {
            Editor editor = Editor.CreateEditor(building);
            editor.DrawDefaultInspector();

            if (GUILayout.Button("Save Changes"))
            {
                EditorDataMapper.CardDataToJson((GraphicCardList)JsonDataMap[DataType.GraphicCardData], (List<GraphicCard>)gameDataMap[DataType.GraphicCardData]);
                DataLoader.SaveData<GraphicCardList>(DataType.GraphicCardData, (GraphicCardList)JsonDataMap[DataType.GraphicCardData]);
                Logger.Log("save card data!");

                var jsonData = DataLoader.LoadData<BuildingEntryList>(DataType.BuildingData);
                var temp = EditorDataMapper.BuildingJsonToData(jsonData);
                gameDataMap[DataType.BuildingData] = (List<Building>)temp.buildings;
            }
        }
    }
    void DrawCardsTab()
    {

        var cards = (List<GraphicCard>)gameDataMap[DataType.GraphicCardData];
        SelectedCardIndex = EditorGUILayout.Popup(SelectedCardIndex, cards.Select(e => e.Name).ToArray());
        var selectedCard = cards[SelectedCardIndex];
        if (selectedCard != null)
        {
            Editor editor = Editor.CreateEditor(selectedCard);
            editor.DrawDefaultInspector();

            if (GUILayout.Button("Save Changes"))
            {
                EditorDataMapper.CardDataToJson((GraphicCardList)JsonDataMap[DataType.GraphicCardData], (List<GraphicCard>)gameDataMap[DataType.GraphicCardData]);
                DataLoader.SaveData<GraphicCardList>(DataType.GraphicCardData, (GraphicCardList)JsonDataMap[DataType.GraphicCardData]);
                Logger.Log("save card data!");
                //update list
                var jsonData = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
                var temp = EditorDataMapper.CardJsonToData(jsonData);
                gameDataMap[DataType.GraphicCardData] = (List<GraphicCard>)temp.cards;
            }
        }
    }

    void DrawPlayerTab()
    {

        var player = (Player)gameDataMap[DataType.PlayerData];

        if (player != null)
        {
            Editor editor = Editor.CreateEditor(player);
            editor.DrawDefaultInspector();

            if (GUILayout.Button("Save Changes"))
            {
                EditorDataMapper.PlayerDataToJson((PlayerEntry)JsonDataMap[DataType.PlayerData], (Player)gameDataMap[DataType.PlayerData]);
                DataLoader.SaveData<PlayerEntry>(DataType.PlayerData, (PlayerEntry)JsonDataMap[DataType.PlayerData]);
                Logger.Log("save card data!");
                //update list
                var jsonData = DataLoader.LoadData<PlayerEntry>(DataType.PlayerData);
                var temp = EditorDataMapper.PlayerJsonToData(jsonData);
                gameDataMap[DataType.PlayerData] = (Player)temp;
            }
        }
    }
}