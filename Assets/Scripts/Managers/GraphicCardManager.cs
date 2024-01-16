using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using System.Linq;
public class GraphicCardManager : MonoBehaviour
{

    public static GraphicCardManager _instance;

    // public List<GameObject> Cards;
    public List<GraphicCard> Cards = new List<GraphicCard>();

    private void Start()
    {
        _instance = this;
        // since graphic card it's not gameobject need to init in the room, so we dont have to setup gameobject
        // when we init cards;

        // decode json to List
        var dataList = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        dataList.GraphicCards.ForEach(e =>
        {

            var card = new GraphicCard();
            card.Name = e.Name;
            card.Id = e.Id;
            card.IsLocked = e.IsLocked;
            card.PerSecondEarn = e.PerSecondEarn;
            card.Price = e.Price;
            card.PerSecondLoseVolt = e.PerSecondLoseVolt;
            card.Quantity = e.Quantity;
            //deal with icon 
            card.Icon = UnityEngine.Resources.Load<Sprite>(Paths.ArtworkFolderPath + e.ImageSource.Path);
            Logger.Log("[GraphicCardManager]: loading card " + e.Name);
            Logger.Log("[GraphicCardManager]: card icon is " + card.Icon);
            Cards.Add(card);
        });
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