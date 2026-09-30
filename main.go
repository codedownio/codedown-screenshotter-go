
package main

import (
	"context"
	"encoding/json"
	"errors"
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
	// Filter out "--" from command line arguments before parsing.
	// This is because we need to do this for reliable parsing when screenshotting
	// on Electron, and this needs to be compatible with the same command line args.
	filteredArgs := make([]string, 0, len(os.Args))
	for _, arg := range os.Args {
		if arg != "--" {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	os.Args = filteredArgs

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

	noHeadless := flag.Bool("no-headless", false, "Pass headless flag to chromedp")

	tmpDir := flag.String("tmp-dir", "", "Temporary directory to use for chromedp")

	destFile := flag.String("dest-file", "screenshot.png", "Destination file to write")
	artifactFile := flag.String("artifact-file", "", "If set, also write the page's preview artifact JSON here. The page supplies it through window.codedownCapturePreview().")

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
	if !*noHeadless {
		options = append(options, chromedp.Flag("headless", true))
	}
	if *tmpDir != "" {
		options = append(options, chromedp.UserDataDir(*tmpDir))
	}

	// Everything that owns the browser lives in a function whose deferred cancels can run.
	// os.Exit skips defers, so exiting from here would leave Chrome running.
	if !screenshot(options, screenshotRequest{
		url:          *url,
		cookieName:   *cookieName,
		cookieValue:  *cookieValue,
		cookieDomain: *cookieDomain,
		width:        *width,
		height:       *height,
		quality:      *quality,
		timeoutMs:    *timeoutMilliseconds,
		destFile:     *destFile,
		artifactFile: *artifactFile,
		debugChrome:  *debugChrome,
	}) {
		os.Exit(1)
	}
}

type screenshotRequest struct {
	url          string
	cookieName   string
	cookieValue  string
	cookieDomain string
	width        int
	height       int
	quality      int
	timeoutMs    int
	destFile     string
	artifactFile string
	debugChrome  bool
}

// How much longer than -timeout-ms the run as a whole is allowed to take, so that a page which
// stalls gets reported by the step that stalled before the overall deadline fires.
const timeoutSlack = 10 * time.Second

func screenshot(options []chromedp.ExecAllocatorOption, req screenshotRequest) bool {
	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer acancel()

	var ctx context.Context
	var cancel context.CancelFunc
	if req.debugChrome {
		ctx, cancel = chromedp.NewContext(actx, chromedp.WithDebugf(log.Printf))
	} else {
		ctx, cancel = chromedp.NewContext(actx)
	}
	defer cancel()

	// A page that never resolves previewReady would otherwise keep this process, and the browser
	// it owns, alive forever: the per-evaluate timeout does not interrupt an awaited promise.
	if req.timeoutMs > 0 {
		var tcancel context.CancelFunc
		ctx, tcancel = context.WithTimeout(ctx, (time.Duration(req.timeoutMs) * time.Millisecond) + timeoutSlack)
		defer tcancel()
	}

	var previewRes bool
	var buf []byte
	var artifact string
	if err := chromedp.Run(ctx, fullScreenshot(
		req.cookieName, req.cookieValue, req.cookieDomain,
		req.url,
		&previewRes,
		req.quality,
		req.timeoutMs,
		req.width,
		req.height,
		&buf,
		req.artifactFile != "",
		&artifact,
	)); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Errorf("Timed out after %dms waiting for %s", req.timeoutMs, req.url)
		} else {
			log.Error(err)
		}
		return false
	}

	if err := os.WriteFile(req.destFile, buf, 0o644); err != nil {
		log.Error(err)
		return false
	}

	log.Printf("Wrote %s", req.destFile)

	// The screenshot is already written and still good, so an artifact failure is reported
	// through the exit code rather than discarding the run.
	return req.artifactFile == "" || writeArtifact(req.artifactFile, artifact)
}

// The page reports its artifact as a JSON string, so that a missing hook and a capture that
// threw are both values rather than evaluation failures that would lose the screenshot.
func writeArtifact(path string, artifact string) bool {
	if artifact == "" || artifact == "null" {
		log.Error("The page did not provide window.codedownCapturePreview; no artifact written")
		return false
	}

	var probe struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(artifact), &probe); err != nil {
		log.Errorf("The page's artifact was not valid JSON: %v", err)
		return false
	}
	if probe.Error != "" {
		log.Errorf("Preview capture failed in the page: %s", probe.Error)
		return false
	}

	if err := os.WriteFile(path, []byte(artifact), 0o644); err != nil {
		log.Errorf("Could not write %s: %v", path, err)
		return false
	}
	log.Printf("Wrote %s (%d bytes)", path, len(artifact))
	return true
}

func fullScreenshot(
	cookieName string,
	cookieValue string,
	cookieDomain string,

	urlstr string,

	previewRes *bool,

	quality int,
	timeoutMilliseconds int,
	width int,
	height int,

	res *[]byte,

	wantArtifact bool,
	artifact *string,
) chromedp.Tasks {
	var actions chromedp.Tasks

	actions = append(actions, chromedp.EmulateViewport(int64(width), int64(height)))

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

	// Before the screenshot: the artifact has to be serialized while the viewport is still the
	// one that was asked for, and FullScreenshot resizes it.
	if wantArtifact {
		actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
			log.Debug("Capturing preview artifact")
			return nil
		}))

		actions = append(actions, chromedp.Evaluate(
			`(window["codedownCapturePreview"] ? window["codedownCapturePreview"]() : Promise.resolve(null))
			   .then(function (r) { return JSON.stringify(r === undefined ? null : r); },
			         function (e) { return JSON.stringify({error: String((e && e.stack) || e)}); });`,
			artifact,
			func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
				var ret = p.WithAwaitPromise(true)
				if timeoutMilliseconds > 0 {
					return ret.WithTimeout(runtime.TimeDelta((timeoutMilliseconds)))
				} else {
					return ret
				}
			}))
	}

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		log.Debug("Taking screenshot")
		return nil
	}))

	actions = append(actions, chromedp.FullScreenshot(res, quality))

	return actions
}
