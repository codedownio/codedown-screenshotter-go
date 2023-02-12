
package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
)

func main() {
	url := flag.String("url", "", "URL to screenshot")
	chromePath := flag.String("chrome-path", "", "Path to chrome or headless-shell executable")

	width := flag.Int("width", 850, "Viewport width")
	height := flag.Int("height", 1000, "Viewport height")

	quality := flag.Int("quality", 90, "PNG quality (0-100)")

	cookieName := flag.String("cookieName", "", "Cookie name")
	cookieValue := flag.String("cookieValue", "", "Cookie value")

	flag.Parse()

	if *url == "" {
		log.Fatal("-url is required")
	}

	headers := make(map[string]interface{})
	if *cookieName != "" && *cookieValue != "" {
		headers["Cookie"] = *cookieName + "=" + *cookieValue
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
	if err := chromedp.Run(ctx, fullScreenshot(*url, *quality, &buf, &headers)); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("screenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	log.Printf("Wrote screenshot.png")
}

func fullScreenshot(
	urlstr string,
	quality int,
	res *[]byte,
	headers *map[string]interface{},
) chromedp.Tasks {
	var actions chromedp.Tasks

	if len(*headers) > 0 {
		actions = append(actions, network.Enable(), network.SetExtraHTTPHeaders(network.Headers(*headers)))
	}

	actions = append(actions, chromedp.Navigate(urlstr))
	actions = append(actions, chromedp.FullScreenshot(res, quality))

	return actions
}
