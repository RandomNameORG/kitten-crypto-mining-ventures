using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using static GraphicCardItem;

public class CardGTX1060 : MonoBehaviour
{
    public GraphicCardItem GTX1060;
    // Start is called before the first frame update
    void Start()
    {
        
    }

    public void updateCard(){
        Transform canvasTransform = GameObject.Find("Canvas").transform;
        GameObject cardPrefab = Resources.Load<GameObject>("GTX1060");
        Instantiate(cardPrefab,canvasTransform);
    }
}
