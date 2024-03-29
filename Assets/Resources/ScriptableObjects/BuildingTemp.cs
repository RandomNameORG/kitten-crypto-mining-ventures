﻿using UnityEngine;
using UnityEngine.UI;
using System.Collections;
using System.Collections.Generic;
using DG.Tweening;
using TMPro;

[CreateAssetMenu(fileName = "NewBuilding", menuName = "ScriptableObjects/Buildings")]
public class BuildingTEMP : ScriptableObject
{
    public string Id;
    public string Name;
    public long Capacity;
    public List<GeneralEvent> Events;
    public List<GraphicCardItem> Cards;
    public double EventHappenProbs;
    public long MoneyPerSecond;
    public List<Alternator> alts;
    public long MaxVolt;
    public long VoltPerSecond;
    public Text voltText;

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
            Logger.LogError("Card not found in " + Id);
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
                Logger.LogError("Power Failure in " + Id);
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
            Logger.LogError("Alternator not found in " + Id);
            return false;
        }


    }




}

