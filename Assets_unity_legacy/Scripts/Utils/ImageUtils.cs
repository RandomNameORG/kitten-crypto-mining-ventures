using System.Collections;
using TMPro;
using UnityEngine;
using UnityEngine.UI;


/// <summary>
/// Utils method class for image processing
/// </summary>
public static class ImageUtils
{
    //Set image transparency 
    public static void SetTransparency(GameObject obj, float alpha)
    {
        var image = obj.GetComponent<Image>();
        if (image != null)
        {
            Color color = image.color;
            color.a = alpha;
            image.color = color;
        }
        else
        {
            Logger.LogError("Image component not found on " + obj.name);
        }
    }

    /// <summary>
    /// The Text comp we always using is //TODO TextMeshPro
    /// </summary>
    /// <param name="obj"></param>
    /// <param name="alpha"></param>
    public static void SetTextTransparency(GameObject obj, float alpha)
    {
        var text = obj.GetComponent<TextMeshProUGUI>();
        if (text != null)
        {
            Color color = text.color;
            color.a = alpha;
            text.color = color;
        }
        else
        {
            Logger.LogError("text component not found on " + obj.name);
        }
    }


    private static IEnumerator ImageFade(GameObject obj, float fadeDuration, float startAlpha, float endAlpha)
    {
        var imageToFade = obj.GetComponent<Image>();
        float elapsedTime = 0;
        Color originalColor = imageToFade.color;
        while (elapsedTime < fadeDuration)
        {
            elapsedTime += Time.deltaTime;
            float alpha = Mathf.Lerp(startAlpha, endAlpha, elapsedTime / fadeDuration);
            imageToFade.color = new Color(originalColor.r, originalColor.g, originalColor.b, alpha);
            yield return null;
        }

        imageToFade.color = new Color(originalColor.r, originalColor.g, originalColor.b, endAlpha);
    }




    //blow if image constant fade in fade out method
    public static IEnumerator ImageFadeIn(GameObject obj, float fadeDuration)
    {
        var imageToFade = obj.GetComponent<Image>();
        var startAlpha = imageToFade.color.a;
        return ImageFade(obj, fadeDuration, startAlpha, 1f);

    }
    public static IEnumerator ImageFadeOut(GameObject obj, float fadeDuration)
    {
        var imageToFade = obj.GetComponent<Image>();
        var startAlpha = imageToFade.color.a;
        return ImageFade(obj, fadeDuration, startAlpha, 0f);
    }

    private static IEnumerator TextFade(GameObject obj, float fadeDuration, float startAlpha, float endAlpha)
    {
        var imageToFade = obj.GetComponent<TextMeshProUGUI>();
        float elapsedTime = 0;
        Color originalColor = imageToFade.color;
        while (elapsedTime < fadeDuration)
        {
            elapsedTime += Time.deltaTime;
            float alpha = Mathf.Lerp(startAlpha, endAlpha, elapsedTime / fadeDuration);
            imageToFade.color = new Color(originalColor.r, originalColor.g, originalColor.b, alpha);
            yield return null;
        }

        imageToFade.color = new Color(originalColor.r, originalColor.g, originalColor.b, endAlpha);
    }
    public static IEnumerator TextFadeOut(GameObject obj, float fadeDuration)
    {
        var textToFade = obj.GetComponent<TextMeshProUGUI>();
        float startAlpha = textToFade.color.a;
        return TextFade(obj, fadeDuration, startAlpha, 0f);
    }

    public static IEnumerator TextFadeIn(GameObject obj, float fadeDuration)
    {
        var textToFade = obj.GetComponent<TextMeshProUGUI>();
        float startAlpha = textToFade.color.a;
        return TextFade(obj, fadeDuration, startAlpha, 1f);
    }


    // public static IEnumerator FadeSequence(GameObject obj, float fadeDuration)
    // {
    //     yield return StartCoroutine(ImageFadeIn(obj, fadeDuration / 2));
    //     yield return new WaitForSeconds(fadeDuration); // wait for second
    //     yield return StartCoroutine(ImageFadeOut(obj, fadeDuration / 2));
    // }
}