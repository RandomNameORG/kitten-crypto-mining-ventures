using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class KeyBindingManager : MonoBehaviour
{
    // Start is called before the first frame update
    public GameObject Pane;
    void Start()
    {
        Pane.SetActive(false);
    }

    // Update is called once per frame
    void Update()
    {
        if(Input.GetKeyDown(KeyCode.Escape)) {
            try
            {
                var win = GameObject.Find("SettingsWindow");
                win.SetActive(false);
            } catch
            {
                if (Pane.activeSelf)
                {
                    TimeUtils.ResumeGame();
                }
                else
                {
                    TimeUtils.PauseGame();
                }

                Pane.SetActive(!Pane.activeSelf);
            }
        }
    }
}
