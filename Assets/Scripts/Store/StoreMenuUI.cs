using System.Collections;
using System.Linq;
using UnityEngine;
using static UnityEditor.Progress;

public class StoreMenuUI : MonoBehaviour
{
    // Start is called before the first frame update
    public StoreItemSlot[] Slots;
    ItemManager _itemManager;
    public GameObject originSlot;
    void Start()
    {
        _itemManager = ItemManager.Instance;
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
        GraphicCardItem[] items = _itemManager.GraphicCardItems;

        for (int i = 0; i < _itemManager.GraphicCardItems.Length; i++)
        {
            Slots[i].AddItem(items[i]);
        }
    }
}
