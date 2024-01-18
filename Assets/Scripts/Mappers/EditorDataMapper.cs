using System.Collections.Generic;
using UnityEngine;
using System.Linq;

/// <summary>
/// this data mapper class don't care relationship between data.
/// </summary>
public static class EditorDataMapper
{

    public class BuildingDTO
    {
        public List<GameObject> Buildings = new();
        public List<Building> buildings = new List<Building>();
    }

    /*
    Here is Building class mapper methods
    */


    /// <summary>
    /// Mapper BuildingEntry json data class to GameObject class Building
    /// </summary>
    /// <param name="jsonData"></param>
    /// <returns></returns>
    public static BuildingDTO BuildingJsonToData(BuildingEntryList jsonData)
    {

        //its okay in this calss
        var cardsData = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        List<GraphicCard> cards = CardJsonToData(cardsData).cards;

        BuildingDTO res = new();
        jsonData.Buildings.ForEach(e =>
        {
            var obj = new GameObject(e.Name);
            //create building comp
            obj.AddComponent<Building>();
            var building = obj.GetComponent<Building>();
            building.Id = e.Id.ToString(); // Assuming  convert the int Id to string
            building.Name = e.Name;
            building.Capacity = e.MaxCardNum;
            building.MaxVolt = e.MaxVolt; // Assuming Capacity is equivalent to MaxVolt
            building.Events = new List<GeneralEvent>(e.Events);
            building.Cards = e.CardSlots.Select(cs =>
            {
                return cards.Find(card => cs.Id == card.Id);
            }).ToList();
            building.EventHappenProbs = e.ProbabilityOfBeingAttacked;
            building.MoneyPerSecond = e.MoneyPerSecond;
            building.Alts = new List<Alternator>(e.alts);
            building.VoltPerSecond = e.VoltPerSecond;

            res.buildings.Add(building);
            res.Buildings.Add(obj);

            //here we have to do something to building house
            //make it exist in the world
            obj.transform.SetPositionAndRotation(new Vector3(0, 0, 0), Quaternion.identity);
            obj.transform.localScale = new Vector3(1, 1, 1);
        });
        return res;
    }

    public static void BuildingDataToJson(BuildingEntryList jsonData, List<GameObject> buildings)
    {
        for (int i = 0; i < jsonData.Buildings.Count; i++)
        {
            BuildingEntry e = jsonData.Buildings[i];
            var building = buildings[i].GetComponent<Building>();
            e.Id = "1";
            e.Name = building.Name;
            e.MaxCardNum = building.Capacity; // Assuming MaxCardNum is equivalent to Capacity
            e.MaxVolt = building.MaxVolt; // Assuming this assignment logic remains the same
            // Assuming e has a CardSlots property that can be assigned from building.Cards
            e.CardSlots = building.Cards.Select(gc =>
            {
                return new GraphicCardReference { Id = gc.Id, Name = gc.Name }; // Assuming CardSlot has an Id property and you can create new instances like this
            }).ToList();
            e.ProbabilityOfBeingAttacked = building.EventHappenProbs;
            e.MoneyPerSecond = building.MoneyPerSecond;
            e.VoltPerSecond = building.VoltPerSecond;
        }
    }

    public class CardDTO
    {
        public List<GameObject> Cards = new();
        public List<GraphicCard> cards = new();
    }
    public static CardDTO CardJsonToData(GraphicCardList jsonData)
    {

        CardDTO res = new();
        jsonData.GraphicCards.ForEach(e =>
        {


            var obj = new GameObject(e.Name);
            //create building comp
            obj.AddComponent<GraphicCard>();
            var card = obj.GetComponent<GraphicCard>();
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
            res.cards.Add(card);
            res.Cards.Add(obj);
        });
        return res;
    }

    public static void CardDataToJson(GraphicCardList jsonData, List<GraphicCard> cards)
    {
        for (int i = 0; i < jsonData.GraphicCards.Count; i++)
        {
            var card = cards[i];
            GraphicCardEntry e = jsonData.GraphicCards[i];
            e.IsLocked = card.IsLocked;
            e.PerSecondEarn = card.PerSecondEarn;
            e.Price = card.Price;
            e.PerSecondLoseVolt = card.PerSecondLoseVolt;
            e.Quantity = card.Quantity;
        }
    }

    public static Player PlayerJsonToData(PlayerEntry jsonData)
    {

        var buildingData = DataLoader.LoadData<BuildingEntryList>(DataType.BuildingData);
        List<Building> buildings = BuildingJsonToData(buildingData).buildings;

        GameObject obj = new GameObject("player");
        obj.AddComponent<Player>();
        Player res = obj.GetComponent<Player>();

        Logger.Log(jsonData.ToString());
        res.Name = jsonData.Name;
        res.TechPoint = jsonData.TechPoint;
        res.Money = jsonData.Money;
        res.TotalCardNum = jsonData.TotalCardNum;
        var tempBuild = buildings.Find(i => i.Id == jsonData.CurrBuildingAt.Id);
        res.CurrBuildingAt = tempBuild;
        res.Buildings = buildings;
        return res;
    }

    public static void PlayerDataToJson(PlayerEntry jsonData)
    {

        List<BuildingReference> buildingRefs = new();
        jsonData.BuildingsRef.ForEach(item =>
        {
            buildingRefs.Add(new BuildingReference
            {
                Id = item.Id,
                Name = item.Name
            });
        });
        PlayerEntry data = new PlayerEntry
        {
            Name = jsonData.Name,
            TechPoint = jsonData.TechPoint,
            Money = jsonData.Money,
            TotalCardNum = jsonData.TotalCardNum,
            CurrBuildingAt = new BuildingReference
            {
                Id = jsonData.CurrBuildingAt.Id,
                Name = jsonData.CurrBuildingAt.Name
            },
            BuildingsRef = buildingRefs
        };

    }

    public static void PlayerDataToJson(PlayerEntry jsonData, Player player)
    {

        jsonData.Name = player.Name;
        jsonData.TechPoint = player.TechPoint;
        jsonData.Money = player.Money;
        jsonData.TotalCardNum = player.TotalCardNum;
        jsonData.CurrBuildingAt = new BuildingReference
        {
            Id = player.CurrBuildingAt.Id,
            Name = player.CurrBuildingAt.Name
        };
        jsonData.BuildingsRef = player.Buildings.Select(building => new BuildingReference()
        {
            Id = building.Id,
            Name = building.Name
        }).ToList();

    }

}