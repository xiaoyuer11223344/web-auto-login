package tests

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func Test_image_ocr_verify(t *testing.T) {
	apiURL := "http://120.26.57.12:8000/ocr"
	imagePath := "/Users/lingchi/study-something/golang/web-auto-login/tests/3.png"
	imageData, err := ioutil.ReadFile(imagePath)
	_ = imageData
	if err != nil {
		panic(err)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	//base64Image := "iVBORw0KGgoAAAANSUhEUgAAAFAAAAAUCAMAAAAtBkrlAAACZ1BMVEX///8AAAAkDYUsESaVUuhZDqmiAJ+vgcsEG4VAg8wVJX8cPdCahTEWrWZy1gk4aKCU1bsru3Vvc2JUCgeC+A/RS8PK1nUhF4f96kzixIfWqwgvhiB1qKAbaiBjGHhY/eoQlvN6wVijTmpapugPT7zYQZZ0QvadQF8deBYyS43kZ11twkrhZjrNwcOOtyfg8Le3+Y75dYYifQQl200eX8yf6a7FwXQmuEdE4id0YBAIqy6QDC62NN+E09qUYQyTsZmjet8UIx0AFU2t/ogPQlqaqppwVcfp4JTVaXW5uPxXQJU1s8KlYKfCEfYeGSMsk/nu+Ja1SmjXzA9ajmzNxpWKDuPyVTK05IBaul3W8RcpZVktmuWELGp4nlNQhdI7RP5SdpmJMMLU9oTYrjiq/pKtRwo1gLhX/xXvyR3hLt2jF6pFfdMFxME4WOqZxYza6kJe1lNIxX5jY7u79nuckVkL7aqK6SZeb2K4AZLhHYYkJzGAmQjKE0iRACuUzxn+YKL1Mgy5KZ2qUn2HXvtXWqSHZ7EU1jeWFzpVKYBWxIqdeAN77R0Qr+9FL4SYl3G37zODFrJupg6HbbUIpoTczxmEnfz8mHdJQYdzBWY4rtsI8Wnx91NNu79du7RQL54SYV2Yymc/kuZHym20J786aLUJpPdOJzfLKs6rsnSaVUOADzGS3AvtmO7oq3tahCb89hbNZzPrrT03g/ZH2rD75tZ6+TsSV43D6WjiTIbCZHc/aHclAYvk34FIsw8bQmPyLTOmuqwKsra7xmNWx0gBg/eYZRazlUHdc3yrPUPvqu+VYjCSzm2PRyju015fAAAACXBIWXMAAA7EAAAOxAGVKw4bAAABP0lEQVQ4jZWSu0pDQRCGZ3an1S5vYClIGrExBJ9A0MRHiGilWIRgKdiJVoLYWAhWyQOIlYKCPpR7v7k75/iTM7uZ2fn2z+QA1HXSyEctOk/81aBd2upHOO08sdRh+xVgvR8REM1y3irw2q20NfrQV9IDl7DGX/AQ2uYVf7Zk42aH1fvQ5rsw2wQW+kpKbyq0huA2O2o9TGvXdpkwQKICk92T+MlGGU9+6fBi91IJ4JHQKRy7gTzDAUtJIyJKHR7FrvJPuesxQ6EkJSKR9Fbc6mK0Fh8G+G6hiALI/nrpMIUZnftxF6nPLeMxWh3SmPwcZFF+6iZUqcdqoHZb5zb1CaNKduNCQxVThNQ/uVXt6fdTQUWSi9yV/rqvnm+AKw4zLROG+Zzn3uSM93vGezVMkVkF+OgzhwOGCeYNqNUS6C+JFBi1WsLHSAAAAABJRU5ErkJggg=="

	data := url.Values{}
	data.Set("image", base64Image)
	data.Set("probability", "false")
	data.Set("png_fix", "false")

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(body))
}
