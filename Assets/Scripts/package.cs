using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using static StoreItemSlot;

public class package : MonoBehaviour
{
    public Dictionary<Object, int> items = new Dictionary<Object, int>();
    // Start is called before the first frame update
    public package( Dictionary<Object, int> items)
    {
        this.items = items;
        delivery();
    }
    void Start()
    {
        
    }

    // Update is called once per frame
    void delivery()
    {
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        GameObject cardPrefab = Resources.Load<GameObject>("package");
        Instantiate(cardPrefab, canvasTransform.position, Quaternion.identity, canvasTransform);
        
        foreach (var item in items)
        {
            Debug.Log(item.Key + " " +  item.Value);
        }

    }
}
