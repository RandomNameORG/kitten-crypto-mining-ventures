using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class SizeManager : MonoBehaviour
{

    private BuildingEntry BuildingIfo;
    public BuildingManager _instance = BuildingManager._instance;
    public GameObject Room ;
    public GameObject Restriction;
    public Player Player; //delete this
    private string ID;

    public void Start()
    {
        // 创建新的 GameObject
        GameObject myNewGameObject = new GameObject();
        myNewGameObject.name = "Dynamic Game Object";

        // 添加 SpriteRenderer 组件
        SpriteRenderer spriteRenderer = myNewGameObject.AddComponent<SpriteRenderer>();

        // 加载美术资源 (replace "YourArtworkPath" with the actual path)
        Sprite artworkSprite = UnityEngine.Resources.Load<Sprite>("Artwork/Sprite-0001");

        //Room = (GameObject)UnityEngine.Resources.Load("Artwork/Sprite-0001");
        Debug.Log(Room.ToString());
        ID = PlayerManager._instance.CurPlayer.CurrBuildingAt.Id;
        // 设置 SpriteRenderer 的 Sprite
            spriteRenderer.sprite = artworkSprite;

            // 修改大小 (replace with your desired size)
            myNewGameObject.transform.position = new Vector3(0, 0, 1f);
            myNewGameObject.transform.localScale = new Vector3(1000, 1000, 1f);
    }
    public void Update()
    {
        RoomSize(ID);
    }
    public void RoomSize(string ID)
    {
        BuildingIfo = _instance.FindBuildingEntryById(ID);
        // generate Room at original point with scale
        Room.transform.localPosition = new Vector3(0,0, 3);
        Room.transform.localScale = new Vector3(BuildingIfo.GridSize, BuildingIfo.GridSize, BuildingIfo.GridSize);

        // generate resctriciton at original point with scale
        Restriction.transform.position = new Vector3(0, 0, -5);
        
        float size = (float)(0.002 * BuildingIfo.GridSize * BuildingIfo.GridSize - 1.7 * BuildingIfo.GridSize + 875);

        Restriction.transform.localScale = new Vector3(size, size, size);
    }
}
