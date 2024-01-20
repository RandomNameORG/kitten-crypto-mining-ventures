using System.Collections.Generic;
using UnityEngine;
using System.Linq;
public static class DataMapper
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
    public static BuildingDTO BuildingsJsonToData(BuildingEntryList jsonData)
    {
        //TODO might causing problem decoupling
        //make this function could works without holding data
        var cardsData = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        List<GraphicCard> cards = CardsJsonToData(cardsData).cards;

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

    public static Building GenerateBuilding(BuildingEntry buildingEntry)
    {

        //TODO might causing problem decoupling
        //make this function could works without holding data
        // var cardsData = DataLoader.LoadData<GraphicCardList>(DataType.GraphicCardData);
        // List<GraphicCard> cards = CardsJsonToData(cardsData).cards;

        var obj = new GameObject("Building");
        //create building comp
        obj.AddComponent<Building>();
        var building = obj.GetComponent<Building>();
        building.Id = buildingEntry.Id.ToString(); // Assuming  convert the int Id to string
        building.Name = buildingEntry.Name;
        building.Capacity = buildingEntry.MaxCardNum;
        building.MaxVolt = buildingEntry.MaxVolt; // Assuming Capacity is equivalent to MaxVolt
        building.Events = new List<GeneralEvent>(buildingEntry.Events);
        building.Cards = buildingEntry.CardSlots.Select(cs =>
        {
            return GenerateCards(cs);
        }).ToList();
        building.EventHappenProbs = buildingEntry.ProbabilityOfBeingAttacked;
        building.MoneyPerSecond = buildingEntry.MoneyPerSecond;
        building.Alts = new List<Alternator>(buildingEntry.alts);
        building.VoltPerSecond = buildingEntry.VoltPerSecond;
        //here we have to do something to building house 
        //TODO here
        obj.transform.SetPositionAndRotation(new Vector3(0, 0, 0), Quaternion.identity);
        obj.transform.localScale = new Vector3(1, 1, 1);
        return building;
    }

    private static GraphicCard GenerateCards(GraphicCardReference cardEntry)
    {
        var e = GraphicCardManager._instance.FindCardById(cardEntry.Id);
        var obj = new GameObject(cardEntry.Name);
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
        return card;
    }
    public static void BuildingsDataToJson(BuildingEntryList jsonData, List<GameObject> buildings)
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
    public static void BuildingDataToJson(BuildingEntry buildingEntry, GameObject buildingObj)
    {
        var building = buildingObj.GetComponent<Building>();
        buildingEntry.Id = "1";
        buildingEntry.Name = building.Name;
        buildingEntry.MaxCardNum = building.Capacity; // Assuming MaxCardNum is equivalent to Capacity
        buildingEntry.MaxVolt = building.MaxVolt; // Assuming this assignment logic remains the same
                                                  // Assuming e has a CardSlots property that can be assigned from building.Cards
        buildingEntry.CardSlots = building.Cards.Select(gc =>
        {
            return new GraphicCardReference { Id = gc.Id, Name = gc.Name }; // Assuming CardSlot has an Id property and you can create new instances like this
        }).ToList();
        buildingEntry.ProbabilityOfBeingAttacked = building.EventHappenProbs;
        buildingEntry.MoneyPerSecond = building.MoneyPerSecond;
        buildingEntry.VoltPerSecond = building.VoltPerSecond;
    }
    public class CardDTO
    {
        public List<GameObject> Cards = new();
        public List<GraphicCard> cards = new();
    }
    public static CardDTO CardsJsonToData(GraphicCardList jsonData)
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
            res.cards.Add(card);
            res.Cards.Add(obj);
        });
        return res;
    }

    public static void CardsDataToJson(GraphicCardList jsonData, List<GraphicCard> cards)
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

        //TODO might causing problem decoupling
        //make this function could works without holding data
        // var buildingData = DataLoader.LoadData<BuildingEntryList>(DataType.BuildingData);
        // List<Building> buildings = BuildingJsonToData(buildingData).buildings;

        GameObject obj = new GameObject("player");
        obj.AddComponent<Player>();
        Player res = obj.GetComponent<Player>();

        Logger.Log(jsonData.ToString());
        res.Name = jsonData.Name;
        res.TechPoint = jsonData.TechPoint;
        res.Money = jsonData.Money;
        res.TotalCardNum = jsonData.TotalCardNum;
        var manager = BuildingManager._instance;
        var tempBuild = manager.FindBuildingById(jsonData.CurrBuildingAt.Id);
        res.CurrBuildingAt = tempBuild;
        res.Buildings = jsonData.BuildingsRef.Select(e => manager.FindBuildingById(e.Id)).ToList();
        return res;
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