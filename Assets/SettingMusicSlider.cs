using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class SettingMusicSlider : MonoBehaviour
{
    [SerializeField]private Slider Slider;
    [SerializeField]private TextMeshProUGUI SliderText;
    // Start is called before the first frame update
    void Start()
    {
        Slider.onValueChanged.AddListener((e) =>
        {
            SliderText.text = e.ToString();
        });
    }

    // Update is called once per frame
    void Update()
    {
        
    }
}
