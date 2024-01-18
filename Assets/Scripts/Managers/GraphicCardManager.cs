using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using System.Linq;
public class GraphicCardManager : MonoBehaviour
{

    public static GraphicCardManager _instance;

    // public List<GameObject> Cards;
    public List<GraphicCard> cards = new List<GraphicCard>();

    //havent use yet
    public List<GameObject> Cards = new();

    private GraphicCardList _card_entries;


    private void Start()
    {

        _instance = this;
        // since graphic card it's not gameobject need to init in the room, so we dont have to setup gameobject
        // when we init cards;

        // decode json to List
        _card_entries = DataManager._instance.GetData<GraphicCardList>(DataType.GraphicCardData);
        var cardDTO = DataMapper.CardJsonToData(_card_entries);
        cards = cardDTO.cards;
        Cards = cardDTO.Cards;

    }

    private void OnApplicationQuit()
    {
        DataMapper.CardDataToJson(_card_entries, cards);
    }

    public GraphicCard FindCardById(string id)
    {
        return cards.FirstOrDefault(card => card.Id == id);
    }

    public GraphicCard FindCardByName(string name)
    {
        return cards.FirstOrDefault(card => card.Name == name);
    }
}