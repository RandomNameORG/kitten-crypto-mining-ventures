
public static class StringUtils
{
    static readonly long NUMBER_CONVERTER = 1000006;
    public static string ConvertMoneyNumToString(long Money) 
    {
        return NUMBER_CONVERTER / Money + "." + NUMBER_CONVERTER % Money;
    }
    
}

