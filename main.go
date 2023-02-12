
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
)

func main() {
	url := flag.String("url", "", "URL to screenshot")
	chromePath := flag.String("chrome-path", "", "Path to chrome or headless-shell executable")

	width := flag.Int("width", 850, "Viewport width")
	height := flag.Int("height", 1000, "Viewport height")

	quality := flag.Int("quality", 95, "PNG quality (0-100)")

	cookieName := flag.String("cookieName", "", "Cookie name")
	cookieValue := flag.String("cookieValue", "", "Cookie value")
	// cookieDomain := flag.String("cookieDomain", "", "Cookie domain")

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

	var previewRes bool
	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(
		&headers,
		*url,
		&previewRes,
		*quality, &buf),
	); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("screenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	log.Printf("Wrote screenshot.png")
}

func fullScreenshot(
	headers *map[string]interface{},

	urlstr string,

	previewRes *bool,

	quality int,
	res *[]byte,
) chromedp.Tasks {
	var actions chromedp.Tasks

	if len(*headers) > 0 {
		actions = append(actions, network.Enable(), network.SetExtraHTTPHeaders(network.Headers(*headers)))
	}

	actions = append(actions, chromedp.Navigate(urlstr))

	actions = append(actions, chromedp.Evaluate(`window["previewReady"];`, &previewRes, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
      return p.WithAwaitPromise(true)
    }))

	actions = append(actions, chromedp.FullScreenshot(res, quality))

	return actions
}
