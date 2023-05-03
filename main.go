
package main

import (
	"context"
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
)

var MyExecAllocatorOptions = [...]chromedp.ExecAllocatorOption{
	chromedp.NoFirstRun,
	chromedp.NoDefaultBrowserCheck,

	// After Puppeteer's default behavior.
	chromedp.Flag("disable-background-networking", true),
	chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
	chromedp.Flag("disable-background-timer-throttling", true),
	chromedp.Flag("disable-backgrounding-occluded-windows", true),
	chromedp.Flag("disable-breakpad", true),
	chromedp.Flag("disable-client-side-phishing-detection", true),
	chromedp.Flag("disable-default-apps", true),
	chromedp.Flag("disable-dev-shm-usage", true),
	chromedp.Flag("disable-extensions", true),
	chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees"),
	chromedp.Flag("disable-hang-monitor", true),
	chromedp.Flag("disable-ipc-flooding-protection", true),
	chromedp.Flag("disable-popup-blocking", true),
	chromedp.Flag("disable-prompt-on-repost", true),
	chromedp.Flag("disable-renderer-backgrounding", true),
	chromedp.Flag("disable-sync", true),
	chromedp.Flag("force-color-profile", "srgb"),
	chromedp.Flag("metrics-recording-only", true),
	chromedp.Flag("safebrowsing-disable-auto-update", true),
	chromedp.Flag("enable-automation", true),
	chromedp.Flag("password-store", "basic"),
	chromedp.Flag("use-mock-keychain", true),
}


func main() {
	url := flag.String("url", "", "URL to screenshot")
	chromePath := flag.String("chrome-path", "", "Path to chrome or headless-shell executable")

	width := flag.Int("width", 850, "Viewport width")
	height := flag.Int("height", 1000, "Viewport height")

	quality := flag.Int("quality", 95, "PNG quality (0-100)")

	timeoutMilliseconds := flag.Int("timeout-ms", 0, "Timeout in milliseconds. Pass 0 to use no timeout.")

	cookieName := flag.String("cookieName", "", "Cookie name")
	cookieValue := flag.String("cookieValue", "", "Cookie value")
	cookieDomain := flag.String("cookieDomain", "", "Cookie domain")

	debug := flag.Bool("debug", false, "Enable debug output")
	debugChrome := flag.Bool("debug-chrome", false, "Enable Chrome debug output")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *url == "" {
		log.Fatal("-url is required")
	}

	options := []chromedp.ExecAllocatorOption{}
	options = append(options, MyExecAllocatorOptions[:]...)
	options = append(options, chromedp.DisableGPU)
	options = append(options, chromedp.WindowSize(*width, *height))
	if *chromePath != "" {
		options = append(options, chromedp.ExecPath(*chromePath))
	}

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer acancel()

	var ctx context.Context
	var cancel context.CancelFunc
	if *debugChrome {
		ctx, cancel = chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(actx)
	}
	defer cancel()

	var previewRes bool
	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot(
		*cookieName, *cookieValue, *cookieDomain,
		*url,
		&previewRes,
		*quality,
		*timeoutMilliseconds,
		&buf,
	)); err != nil {
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
	timeoutMilliseconds int,

	res *[]byte,
) chromedp.Tasks {
	var actions chromedp.Tasks

	if cookieName != "" && cookieValue != "" {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			expires := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
			log.Debug("Setting cookie")
			err := network.SetCookie(cookieName, cookieValue).
				WithExpires(&expires).
				WithDomain(cookieDomain).
				WithHTTPOnly(true).
				Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}))
	}

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		log.Debug("Navigating")
		return nil
	}))

	actions = append(actions, chromedp.Navigate(urlstr))

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		log.Debug("Waiting for previewReady promise")
		return nil
	}))

	actions = append(actions, chromedp.Evaluate(`window["previewReady"];`, &previewRes, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
		var ret = p.WithAwaitPromise(true)
		if timeoutMilliseconds > 0 {
			return ret.WithTimeout(runtime.TimeDelta((timeoutMilliseconds)))
		} else {
			return ret
		}
    }))

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		log.Debug("Taking screenshot")
		return nil
	}))

	actions = append(actions, chromedp.FullScreenshot(res, quality))

	return actions
}
