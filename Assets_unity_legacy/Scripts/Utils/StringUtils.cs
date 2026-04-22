
public static class StringUtils
{
    static readonly long NUMBER_CONVERTER = 1000000;
    public static string ConvertMoneyNumToString(long Money) 
    {
        return Money / NUMBER_CONVERTER + "." + Money % NUMBER_CONVERTER;
    }
    
}

