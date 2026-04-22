using System.Collections;
using System.Collections.Generic;
using UnityEngine;

/// <summary>
/// This class prevent some fixed game object destroy by changing scene
/// </summary>
public class AvoidDestroy : MonoBehaviour
{
    // Start is called before the first frame update
    void Start()
    {
        DontDestroyOnLoad(this.gameObject);
    }
}
