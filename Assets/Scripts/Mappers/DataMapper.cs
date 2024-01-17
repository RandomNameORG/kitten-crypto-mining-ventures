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
    public static BuildingDTO BuildingJsonToData(BuildingEntryList jsonData)
    {

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
                return GraphicCardManager._instance.FindCardById(cs.Id);
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
}