using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using System.Linq;
public class GraphicCardManager : MonoBehaviour
{

    public static GraphicCardManager _instance;

    // public List<GameObject> Cards;
    public List<GraphicCard> Cards = new List<GraphicCard>();

    private GraphicCardList _card_entries;


    private void Start()
    {
        _instance = this;
        // since graphic card it's not gameobject need to init in the room, so we dont have to setup gameobject
        // when we init cards;

        // decode json to List
        _card_entries = DataManager._instance.GetData<BuildingEntryList>(DataType.BuildingData);
        Cards = DataMapper.BuildingJsonToData(_card_entries);
        
    }

    private void OnApplicationQuit()
    {
        //DataMapper.BuildingDataToJson(_building_entries, Buildings);
    }

    public GraphicCard FindCardById(string id)
    {
        return Cards.FirstOrDefault(card => card.Id == id);
    }

    public GraphicCard FindCardByName(string name)
    {
        return Cards.FirstOrDefault(card => card.Name == name);
    }
}