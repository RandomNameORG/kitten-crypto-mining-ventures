using System.Collections;
using System;
using System.Collections.Generic;
using UnityEngine;

[CreateAssetMenu(fileName = "Object", menuName ="ScriptableObjects/OldObject")]
public class ObjectSO : ScriptableObject
{
    
    public List<OldObject> Objects;
}

[Serializable]
public class OldObject{
    [field:SerializeField]
    public string Name {get; private set; }
    [field:SerializeField]
    public int ID {get; private set; }
    [field:SerializeField]
    public Vector2Int Size {get; private set; } = Vector2Int.one;
    [field:SerializeField]
    public GameObject Prefab  {get; private set; }
}

