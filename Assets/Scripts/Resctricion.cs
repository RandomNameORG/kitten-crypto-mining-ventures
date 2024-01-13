using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class Resctricion : MonoBehaviour
{
    private Collider2D myCollider;
    public GameObject buildingSystem; // 指定你要隐藏的物体

    private void Start()
    {
        // 获取物体上的 Collider2D 组件
        myCollider = GetComponent<Collider2D>();

        // 检查是否成功获取 Collider2D
        if (myCollider == null)
        {
            Logger.LogError("Collider2D not found!");
        }
    }

    private void Update()
    {
        // 获取鼠标在屏幕上的位置
        Vector2 mousePosition = Input.mousePosition;

        // 将屏幕坐标转换为世界坐标
        Vector2 worldMousePosition = Camera.main.ScreenToWorldPoint(mousePosition);

        // 检查鼠标位置是否与物体发生碰撞
        bool isColliding = myCollider.OverlapPoint(worldMousePosition);

        // 处理碰撞结果
        if (!isColliding)
        {
            Logger.Log("collidingh");
            // 获取 BuildingSystem 物体上的 Renderer 组件
            buildingSystem.SetActive(false);
        }
        else
        {
            buildingSystem.SetActive(true);
        }
    }
}
