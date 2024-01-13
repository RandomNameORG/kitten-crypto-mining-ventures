using System.Collections;
using System.Collections.Generic;
using System.Linq;
using UnityEngine;
using static UnityEditor.Progress;

public class StorageMenuUI : MonoBehaviour
{
    // Start is called before the first frame update
    public StoreItemSlot[] Slots;
    GraphicCardManager _graphicCardManager;
    public GameObject originSlot;
    void Start()
    {
        _graphicCardManager = GraphicCardManager._instance;
        Transform slotGridTras = transform.Find("StoreItemSlotGrid");
        Slots = transform.GetComponentsInChildren<StoreItemSlot>();
        UpdateUI();
    }

    // Update is called once per frame
    void Update()
    {

    }
    void UpdateUI()
    {
        List<GraphicCard> items = _graphicCardManager.Cards;
        for (int i = 0; i < items.Count; i++)
        {
            Slots[i].AddItem(items[i]);
        }
    }
}
