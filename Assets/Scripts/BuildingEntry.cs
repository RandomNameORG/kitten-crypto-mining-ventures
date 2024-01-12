using System;
using System.Collections;
using System.Collections.Generic;
using AlternatorProject;
using UnityEngine;

[Serializable]
public class BuildingList
{
    public List<BuildingEntry> Buildings = new List<BuildingEntry>();
}

[Serializable]
public class BuildingEntry
{
    public int Id;
    public string Name;
    public int GridSize;
    public Resorces BuildingMaterial;

    public long VoltPerSecond;
    public long MoneyPerSecond; //Money earn persecond
    public long MaxVolt;

    public long MaxCardNum;

    public int ProbabilityOfBeingAttacked;

    public int HeatDissipationLevel;

    public int LocationOfTheBuilding;
    public List<Decoration> Decorations;

    public List<Cat> Cats;

    //this is the card info we need to store
    public List<CardReference> CardSlots;

    //TODO We cannot store entity data for these, each only needs a little bit of data
    //For example, the following three data should not be stored in json
    //using NonSerialized like this for all of them. otherwise, it will apear on json file
    public List<GeneralEvent> Events;
    public List<GraphicCard> Cards;
    public List<Alternator> alts;


    //below is crud methods, have to change when you change the 
    public void AddingGraphicCard(GraphicCard card)
    {
        this.Cards.Add(card);
        this.MoneyPerSecond += card.PerSecondEarn;
        this.VoltPerSecond += card.PerSecondLoseVolt;
    }
    public bool RemoveGraphicCard(GraphicCard card)
    {
        if (Cards.Contains(card))
        {
            Cards.Remove(card);
            MoneyPerSecond -= card.PerSecondEarn;
            VoltPerSecond -= card.PerSecondLoseVolt;
            return true;
        }
        else
        {
            Debug.LogError("Card not found in " + Id);
            return false;
        }
    }
    public int CardSize() { return this.Cards.Count; }

    public void AddingAlternator(Alternator alternator)
    {
        alts.Add(alternator);
        MaxVolt += alternator.MaxVolt;

    }
    public bool RemoveAlternator(Alternator alternator)
    {

        if (alts.Contains(alternator))
        {
            if (VoltPerSecond - alternator.MaxVolt < 0)
            {
                Debug.LogError("Power Failure in " + Id);
            }
            else
            {
                alts.Remove(alternator);
                MaxVolt -= alternator.MaxVolt;
            }
            return true;
        }
        else
        {
            Debug.LogError("Alternator not found in " + Id);
            return false;
        }


    }

}

[Serializable]
public class Decoration
{
    public int ID;
    public GameObject Prefab;
    public float[] Coordinates;
}

[Serializable]
public class Cat
{
    public int ID;
    public string Name;
    //TODO: skills and so on
}

[Serializable]
public class Resorces
{
    public string? RightWallMaterial;
    public string? LeftWallMaterial;

    public string? RightFloorMaterial;
    public string? LeftFloorMaterial;

}
/// <summary>
/// this model stand for speicific position in our grid
/// </summary>
public class GridPosition
{
    int X;
    int Y;
}