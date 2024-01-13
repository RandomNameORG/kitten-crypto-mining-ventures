using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using static StoreItemSlot;

public class Package : MonoBehaviour
{
    public Dictionary<object, int> items = new Dictionary<object, int>();
    // Start is called before the first frame update
    public Package(Dictionary<object, int> items)
    {
        this.items = items;
        Delivery();
    }
    void Start()
    {

    }

    // Update is called once per frame
    void Delivery()
    {
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        GameObject cardPrefab = Resources.Load<GameObject>("package");
        Instantiate(cardPrefab, canvasTransform.position, Quaternion.identity, canvasTransform);

        foreach (var item in items)
        {
            Logger.Log(item.Key + " " + item.Value);
        }

    }
}
