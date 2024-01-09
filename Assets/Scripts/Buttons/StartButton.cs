using UnityEngine.SceneManagement;
using UnityEngine;
using UnityEngine.UI;


public class StartButton : MonoBehaviour
{
    // Start is called before the first frame update
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
        SceneManager.UnloadSceneAsync(SceneManager.GetActiveScene().buildIndex);
        SceneManager.LoadScene("Main");
    }

   
}
