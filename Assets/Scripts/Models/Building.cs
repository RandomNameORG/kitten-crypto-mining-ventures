using UnityEngine;
using UnityEngine.UI;
using System.Collections.Generic;

/// <summary>
/// The object class to attch with gameobject   
/// </summary>
public class Building : MonoBehaviour
{
    public string Id;
    public string Name;
    public long Capacity;
    public List<GeneralEvent> Events;
    public List<GraphicCard> Cards;
    public double EventHappenProbs;
    public long MoneyPerSecond;
    public List<Alternator> Alts;
    public long MaxVolt;
    public long VoltPerSecond;
    public static Text VoltText;


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
            Logger.LogError("Card not found in " + Id);
            return false;
        }
    }
    public int CardSize() { return this.Cards.Count; }

    public void AddingAlternator(Alternator alternator)
    {
        Alts.Add(alternator);
        MaxVolt += alternator.MaxVolt;

    }
    public bool RemoveAlternator(Alternator alternator)
    {

        if (Alts.Contains(alternator))
        {
            if (VoltPerSecond - alternator.MaxVolt < 0)
            {
                Logger.LogError("Power Failure in " + Id);
            }
            else
            {
                Alts.Remove(alternator);
                MaxVolt -= alternator.MaxVolt;
            }
            return true;
        }
        else
        {
            Logger.LogError("Alternator not found in " + Id);
            return false;
        }


    }

}

