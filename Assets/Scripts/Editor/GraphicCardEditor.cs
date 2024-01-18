using UnityEngine;
using UnityEditor;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Linq;

public class GraphicCardEditorWindow : EditorWindow
{
    [MenuItem("Tools/Graphic Card Editor")]
    public static void ShowWindow()
    {
        GetWindow<GraphicCardEditorWindow>("Graphic Card Manager");
    }

    private GraphicCardList _card_entries;
    private List<GraphicCard> Cards;

    private void OnEnable()
    {

        loadData();
        Logger.Log("load graphic cards data done");
    }
    private int CardSelectedIndex = 0;

    void OnGUI()
    {
        CardSelectedIndex = EditorGUILayout.Popup(CardSelectedIndex, Cards.Select(e => e.Name).ToArray());
        var selectedCard = Cards[CardSelectedIndex];
        if (selectedCard != null)
        {
            Editor editor = Editor.CreateEditor(selectedCard);
            editor.DrawDefaultInspector();

            if (GUILayout.Button("Save Changes"))
            {
                saveData();
            }
        }
    }
    void loadData()
    {
        _card_entries = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        Cards = DataMapper.CardJsonToData(_card_entries).cards;
    }
    void saveData()
    {
        DataMapper.CardDataToJson(_card_entries, Cards);
        DataLoader.SaveData<GraphicCardList>(DataType.GraphicCardData, _card_entries);
    }
}

