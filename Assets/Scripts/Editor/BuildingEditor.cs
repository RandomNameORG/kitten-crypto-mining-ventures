using UnityEngine;
using UnityEditor;
using System;
using System.IO;
using UnityEditorInternal;
using System.Linq;
using Unity.Plastic.Antlr3.Runtime;
using Unity.VisualScripting;
using System.Collections.Generic;


public class TestList
{
    public List<GraphicCardReference> data;
}

public class BuildingEditor : EditorWindow
{
    private BuildingEntryList jsonData;
    private List<BuildingEntry> buildings;
    private List<GraphicCardReference> cardTypes;
    private List<List<int>> cardSelectedIndexLists = new List<List<int>>();
    private Vector2 scrollPosition;
    private int selectedIndex = 0;

    private readonly string buildingPath = Application.streamingAssetsPath + "/buildings.json";
    private readonly string cardPath = Application.streamingAssetsPath + "/graphiccards.json";
    [MenuItem("Tools/Building Editor")]
    public static void ShowWindow()
    {
        GetWindow(typeof(BuildingEditor), false, "Building Editor");
    }

    private void OnEnable()
    {
        LoadData();
        InitializeCardSelectionLists();
    }

    private void LoadData()
    {
        jsonData = JsonUtility.FromJson<BuildingEntryList>(File.ReadAllText(buildingPath));
        buildings = jsonData.Buildings;
        cardTypes = LoadCardTypes();
    }

    private void InitializeCardSelectionLists()
    {
        cardSelectedIndexLists = buildings.Select(building =>
            building.CardSlots.Select(cardSlot =>
                cardTypes.FindIndex(cardType => cardType.Id == cardSlot.Id)
            ).ToList()
        ).ToList();
    }

    private List<GraphicCardReference> LoadCardTypes()
    {
        var graphicCardList = JsonUtility.FromJson<GraphicCardList>(File.ReadAllText(cardPath));
        return graphicCardList.GraphicCards
            .Select(card => new GraphicCardReference { Id = card.Id, Name = card.Name })
            .ToList();
    }

    private void OnGUI()
    {
        DrawBuildingSelection();
        DrawBuildingEditor();
    }

    private void DrawBuildingSelection()
    {
        var options = buildings.Select(b => b.Name).Append("Adding New Building").ToArray();
        selectedIndex = EditorGUILayout.Popup("Select Building:", selectedIndex, options);
    }

    private void DrawBuildingEditor()
    {
        if (selectedIndex < buildings.Count)
        {
            scrollPosition = EditorGUILayout.BeginScrollView(scrollPosition);
            DrawBuilding(selectedIndex);
            EditorGUILayout.EndScrollView();
        }
        else
        {
            CreateNewBuilding();
        }
    }

    private void CreateNewBuilding()
    {
        BuildingEntry building = new BuildingEntry()
        {
            Name = "default building"
        };
        buildings.Add(building);
    }

    private void SaveData()
    {
        var json = JsonUtility.ToJson(jsonData, true);
        File.WriteAllText(buildingPath, json);
    }

    /// <summary>
    /// Helper method for build TextField
    /// </summary>
    /// <param name="title"></param>
    /// <param name="data"></param>
    /// <returns></returns>
    private long LongTextField(string title, long data)
    {
        return Int64.Parse(EditorGUILayout.TextField(title, data.ToString()));
    }
    private int IntTextField(string title, int data)
    {
        return Int32.Parse(EditorGUILayout.TextField(title, data.ToString()));
    }

    private double DoubleTextField(string title, double data)
    {
        return Double.Parse(EditorGUILayout.TextField(title, data.ToString()));
    }


    private void DrawBuilding(int index)
    {
        var building = buildings[index];

        DrawBasicBuildingInfo(building);
        DrawDecorations(building);
        DrawGraphicCardSlots(index, building);
        DrawBuildingResources(building);
        DrawBuildingActions(index, building);
    }

    private void DrawBasicBuildingInfo(BuildingEntry building)
    {
        building.Name = EditorGUILayout.TextField("Name:", building.Name);
        building.GridSize = IntTextField("Grid Size:", building.GridSize);
        building.VoltPerSecond = LongTextField("Volt Per Second:", building.VoltPerSecond);
        building.MoneyPerSecond = LongTextField("Money Per Second:", building.MoneyPerSecond);
        building.MaxVolt = LongTextField("Max Volt:", building.MaxVolt);
        building.MaxCardNum = LongTextField("Max Card Num:", building.MaxCardNum);
        building.ProbabilityOfBeingAttacked = DoubleTextField("Probability Of Being Attacked:", building.ProbabilityOfBeingAttacked);
        building.HeatDissipationLevel = IntTextField("Heat Dissipation Level:", building.HeatDissipationLevel);
        building.LocationOfTheBuilding = IntTextField("Location Of The Building:", building.LocationOfTheBuilding);
    }

    private void DrawDecorations(BuildingEntry building)
    {
        EditorGUILayout.LabelField("Decorations:");
        if (building.Decorations == null) building.Decorations = new List<Decoration>();

        for (int i = 0; i < building.Decorations.Count; i++)
        {
            DrawDecorationItem(building.Decorations, i);
        }

        if (GUILayout.Button("Add Decoration"))
        {
            building.Decorations.Add(new Decoration());
        }
    }

    private void DrawDecorationItem(List<Decoration> decorations, int index)
    {
        var decoration = decorations[index];
        EditorGUI.indentLevel++;
        EditorGUILayout.LabelField("Decoration " + (index + 1));
        decoration.Resource.Path = EditorGUILayout.TextField("Resource Path:", decoration.Resource.Path);
        decoration.Coordinates.X = IntTextField("Decro X:", decoration.Coordinates.X);
        decoration.Coordinates.Y = IntTextField("Decro Y:", decoration.Coordinates.Y);
        EditorGUI.indentLevel--;

        if (GUILayout.Button("Remove Decoration"))
        {
            decorations.RemoveAt(index);
        }
    }

    private void DrawGraphicCardSlots(int buildingIndex, BuildingEntry building)
    {
        EditorGUILayout.LabelField("Graphic Cards List:");

        for (int i = 0; i < building.CardSlots.Count; i++)
        {
            DrawGraphicCardSlot(buildingIndex, building, i);
        }

        if (GUILayout.Button("Add Card"))
        {
            var newCardReference = new GraphicCardReference() { Id = "1" };
            building.CardSlots.Add(newCardReference);


            if (cardSelectedIndexLists.Count > buildingIndex)
            {
                cardSelectedIndexLists[buildingIndex].Add(cardTypes.FindIndex(cardType => cardType.Id == newCardReference.Id));
            }
        }
    }

    private void DrawGraphicCardSlot(int buildingIndex, BuildingEntry building, int slotIndex)
    {
        EditorGUI.indentLevel++;
        var slot = building.CardSlots[slotIndex];
        Debug.Log("building carslots length: " + cardSelectedIndexLists[buildingIndex].Count);
        cardSelectedIndexLists[buildingIndex][slotIndex] = EditorGUILayout.Popup("Card:", cardSelectedIndexLists[buildingIndex][slotIndex], cardTypes.Select(c => c.Name).ToArray());

        if (GUILayout.Button("Remove Card"))
        {
            building.CardSlots.RemoveAt(slotIndex);
        }
        EditorGUI.indentLevel--;
    }
    private void DrawBuildingResources(BuildingEntry building)
    {
        EditorGUILayout.LabelField("Resources:");
        EditorGUI.indentLevel++;
        building.BuildingMaterial.LeftFloorMaterial = EditorGUILayout.TextField("Left Floor Source:", building.BuildingMaterial.LeftFloorMaterial);
        building.BuildingMaterial.RightFloorMaterial = EditorGUILayout.TextField("Right Floor Source:", building.BuildingMaterial.RightFloorMaterial);
        building.BuildingMaterial.LeftWallMaterial = EditorGUILayout.TextField("Left Wall Source:", building.BuildingMaterial.LeftWallMaterial);
        building.BuildingMaterial.RightWallMaterial = EditorGUILayout.TextField("Right Wall Source:", building.BuildingMaterial.RightWallMaterial);
        EditorGUI.indentLevel--;
    }
    private void DrawBuildingActions(int index, BuildingEntry building)
    {
        EditorGUILayout.Space();
        EditorGUILayout.BeginHorizontal();
        GUILayout.FlexibleSpace();
        if (GUILayout.Button("Save"))
        {
            SaveData();
            Debug.Log("Building saved: " + building.Name);
        }

        if (GUILayout.Button("Remove Building"))
        {
            if (EditorUtility.DisplayDialog("Delete Confirmation",
                                            "Are you sure you want to delete this building?",
                                            "Yes", "No"))
            {
                buildings.RemoveAt(index);
                SaveData();
                selectedIndex = 0;
            }
        }

        EditorGUILayout.EndHorizontal();
        EditorGUILayout.Space();
    }
}

