using UnityEngine;
using System.Collections;
using System.Collections.Generic;

[CreateAssetMenu(fileName = "NewBuilding", menuName = "ScriptableObjects/Buildings")]
public class Building : ScriptableObject
{
    public string Id;
    public string Name;
    public long Capacity;
    public List<GeneralEvent> Events;
    public List<GraphicCardItem> Cards;
    public double EventHappenProbs;
    public long MoneyPerSecond;
    public long VoltPerSecond;

    public void AddingGraphicCard(GraphicCardItem card)
    {
        this.Cards.Add(card);
        this.MoneyPerSecond += card.PerSecondEarn;
        this.VoltPerSecond += card.PerSecondLoseVolt;
    }
    public bool RemoveGraphicCard(GraphicCardItem card)
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
}

