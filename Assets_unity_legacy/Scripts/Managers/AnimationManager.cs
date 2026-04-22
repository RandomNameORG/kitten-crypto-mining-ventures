using System.Collections;
using System.Collections.Generic;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

/// <summary>
/// Singleton class for manager animation
/// This is General gameobject animation manager
/// such as gameobject fade in-n-out
/// using this to control all the genral animation
/// </summary>
public class AnimationManager : MonoBehaviour
{

    public static AnimationManager _instance;
    //all manager should init at @Start stage
    private void Start()
    {
        _instance = this;
        Logger.Log("animation manager init! " + this);
    }

    /// <summary>
    /// Fade in and out all the GameObjects containing Image or TextMeshProUGUI in this gameobject, include this one
    /// </summary>
    /// <param name="obj">The GameObject to fade</param>
    /// <param name="fadeDuration">Duration of the fade</param>
    /// <param name="waitDuration">Duration to wait between fade in and fade out</param>
    public IEnumerator FadeSequence(GameObject obj, float fadeDuration, float waitDuration)
    {
        // fade in
        yield return StartCoroutine(Fade(obj, fadeDuration, true));

        // wait
        yield return new WaitForSeconds(waitDuration);

        // fade out
        yield return StartCoroutine(Fade(obj, fadeDuration, false));
    }
    private IEnumerator Fade(GameObject obj, float fadeDuration, bool fadeIn)
    {
        // List to keep track of all started coroutines
        List<Coroutine> coroutines = new List<Coroutine>();

        // Check if the current object has an Image or TextMeshProUGUI component,
        // and start the corresponding fade in or fade out coroutine
        Image image = obj.GetComponent<Image>();
        if (image != null)
        {
            Coroutine fadeCoroutine = StartCoroutine(
                fadeIn ? ImageUtils.ImageFadeIn(obj, fadeDuration) : ImageUtils.ImageFadeOut(obj, fadeDuration)
            );
            coroutines.Add(fadeCoroutine);
        }

        TextMeshProUGUI text = obj.GetComponent<TextMeshProUGUI>();
        if (text != null)
        {
            Coroutine fadeCoroutine = StartCoroutine(
                fadeIn ? ImageUtils.TextFadeIn(obj, fadeDuration) : ImageUtils.TextFadeOut(obj, fadeDuration)
            );
            coroutines.Add(fadeCoroutine);
        }

        // Iterate over all child GameObjects and start their fade coroutines
        foreach (Transform child in obj.transform)
        {
            Coroutine childCoroutine = StartCoroutine(Fade(child.gameObject, fadeDuration, fadeIn));
            coroutines.Add(childCoroutine);
        }

        // Wait for all fade coroutines to complete
        foreach (var coroutine in coroutines)
        {
            yield return coroutine;
        }
    }


}