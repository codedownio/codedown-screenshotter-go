
package main

import (
	"context"
	"log"
	"os"

	"github.com/chromedp/chromedp"
)

func main() {
	var url string = `https://brank.as/`
	var width int = 850
	var height int = 850
	var chromePath string = "/usr/bin/google-chrome";

	options := []chromedp.ExecAllocatorOption{}
	options = append(options, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options, chromedp.DisableGPU)
	options = append(options, chromedp.WindowSize(width, height))
	options = append(options, chromedp.ExecPath(chromePath))

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer acancel()
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	var quality int = 90
	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(url, quality, &buf)); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	log.Printf("Wrote fullScreenshot.png")
}

func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}
