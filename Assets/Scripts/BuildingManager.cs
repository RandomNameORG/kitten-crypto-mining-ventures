using UnityEngine;
using System.Collections;
using System.Linq;

public class BuildingManager : MonoBehaviour
{
	public static BuildingManager Instance;
	public Building[] Buildings;
    // Use this for initialization
    private void Awake()
    {
		Instance = this;

        Debug.Log("Building Manager init...");

        Buildings = Utils.GetAllInstance<Building>();
        foreach (Building building in Buildings)
        {
            Debug.Log(building);
        }
        Debug.Log("Building Manager done");
    }
    void Start()
    {

    }

    // Read: Find a Building by its ID
    public Building FindBuildingById(string id)
    {
        return Buildings.FirstOrDefault(building => building.Id.Equals(id));
    }
    // Read: Find a Building by its name
    public Building FindBuildingByName(string name)
    {
        return Buildings.FirstOrDefault(building => building.Name.Equals(name));
    }
}


