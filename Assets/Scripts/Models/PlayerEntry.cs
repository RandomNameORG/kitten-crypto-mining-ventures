using System;
using System.Collections.Generic;

[Serializable]
public class PlayerEntry
{
    public string Name = "Woshishabiyouxi";
    public int TechPoint = 0;
    public long Money = 0;
    public int TotalCardNum = 0;
    //the current building player at;
    public BuildingReference CurrBuildingAt;
    public List<BuildingReference> BuildingsRef = new();
}
