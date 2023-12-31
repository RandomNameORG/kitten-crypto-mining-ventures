using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class KeyBindingManager : MonoBehaviour
{
    // Start is called before the first frame update
    public GameObject Pane;
    void Start()
    {
        Pane = GameObject.Find("EscWindow");
        Pane.SetActive(false);
    }

    // Update is called once per frame
    void Update()
    {
        if(Input.GetKeyDown(KeyCode.Escape)) {
            Pane.SetActive(!Pane.activeSelf);
        }
    }
}
