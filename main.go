
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
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
	cookieDomain := flag.String("cookieDomain", "", "Cookie domain")

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

	var previewRes bool
	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(
		*cookieName, *cookieValue, *cookieDomain,
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
	cookieName string,
	cookieValue string,
	cookieDomain string,

	urlstr string,

	previewRes *bool,

	quality int,
	res *[]byte,
) chromedp.Tasks {
	var actions chromedp.Tasks

	if cookieName != "" && cookieValue != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			err := network.SetCookie(cookieName, cookieValue).
				WithExpires(&expr).
				WithDomain(cookieDomain).
				WithHTTPOnly(true).
				Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}))
	}

	actions = append(actions, chromedp.Navigate(urlstr))

	actions = append(actions, chromedp.Evaluate(`window["previewReady"];`, &previewRes, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
      return p.WithAwaitPromise(true)
    }))

	actions = append(actions, chromedp.FullScreenshot(res, quality))

	return actions
}
