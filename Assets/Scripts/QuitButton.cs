using System.Collections;
using System.Collections.Generic;
using UnityEngine.SceneManagement;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;

public class QuitButton : MonoBehaviour
{
    void Start()
    {
        GetComponent<Button>().onClick.AddListener(ButtonClick);
    }

    // Update is called once per frame
    void Update()
    {
        
    }
    void ButtonClick()
    {
        Application.Quit();//ÍË³öÓÎÏ·
    }
}
