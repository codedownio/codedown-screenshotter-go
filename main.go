
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/chromedp/chromedp"
)

func main() {
	width := flag.Int("width", 850, "Viewport width")
	height := flag.Int("height", 1000, "Viewport height")
	quality := flag.Int("quality", 90, "PNG quality (0-100)")
	chromePath := flag.String("chrome-path", "", "Path to chrome or headless-shell executable")
	url := flag.String("url", "", "URL to screenshot")
	flag.Parse()

	if *url == "" {
		log.Fatal("-url is required")
	}

	options := []chromedp.ExecAllocatorOption{}
	options = append(options, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options, chromedp.DisableGPU)
	options = append(options, chromedp.WindowSize(*width, *height))
	if *chromePath != "" {
		options = append(options, chromedp.ExecPath(*chromePath))
	}

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer acancel()
	ctx, cancel := chromedp.NewContext(actx)
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(*url, *quality, &buf)); err != nil {
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
