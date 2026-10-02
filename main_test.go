package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// These exercise the real browser, so there is nothing to assert without one.
func browserOrSkip(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("CODEDOWN_TEST_CHROME"); path != "" {
		return path
	}
	for _, name := range []string{
		"headless-shell", "chromium-headless-shell", "chromium", "chrome", "google-chrome",
	} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	t.Skip("no browser on PATH; set CODEDOWN_TEST_CHROME to run")
	return ""
}

func servePage(t *testing.T, body string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

// The same options main builds, so a test covers the configuration that actually ships.
func testOptions(t *testing.T) []chromedp.ExecAllocatorOption {
	t.Helper()
	options := []chromedp.ExecAllocatorOption{}
	options = append(options, MyExecAllocatorOptions[:]...)
	options = append(options,
		chromedp.DisableGPU,
		chromedp.WindowSize(680, 800),
		chromedp.ExecPath(browserOrSkip(t)),
		chromedp.Flag("headless", true),
		chromedp.UserDataDir(t.TempDir()),
	)
	return options
}

func request(t *testing.T, url string, timeoutMs int, destName string, wantArtifact bool) screenshotRequest {
	t.Helper()
	dir := t.TempDir()
	req := screenshotRequest{
		url:       url,
		width:     680,
		height:    800,
		quality:   95,
		timeoutMs: timeoutMs,
		destFile:  filepath.Join(dir, destName),
	}
	if wantArtifact {
		req.artifactFile = filepath.Join(dir, "artifact.json")
	}
	return req
}

const stalledPage = `<!DOCTYPE html><html><body>stalled` +
	`<script>window.previewReady = new Promise(function () {});</script></body></html>`

const readyPage = `<!DOCTYPE html><html><body>ready` +
	`<script>window.previewReady = Promise.resolve(true);` +
	`window.codedownCapturePreview = function () {` +
	`  return Promise.resolve({html: "<p>hello</p>", swatches: {byTheme: {}}, width: 680, height: 800});` +
	`};</script></body></html>`

const readyPageNoHook = `<!DOCTYPE html><html><body>ready` +
	`<script>window.previewReady = Promise.resolve(true);</script></body></html>`

// The bug this guards: a page whose previewReady never settles kept the process, and the
// browser it owns, alive indefinitely. Awaiting a promise that never resolves is not
// interrupted by the per-evaluate timeout, so only the run-level deadline ends it.
func TestStalledPreviewPromiseDoesNotOutliveItsDeadline(t *testing.T) {
	options := testOptions(t)
	req := request(t, servePage(t, stalledPage), 3000, "out.png", true)

	started := time.Now()
	ok := screenshot(options, req)
	elapsed := time.Since(started)

	if ok {
		t.Fatal("expected a stalled page to fail the run")
	}
	// Anything quicker than the timeout means the run failed for some other reason -- a browser
	// that would not start, say -- and says nothing about the deadline.
	waited := time.Duration(req.timeoutMs) * time.Millisecond
	if elapsed < waited {
		t.Fatalf("gave up after %v without waiting out the %v timeout", elapsed, waited)
	}
	budget := waited + timeoutSlack + 20*time.Second
	if elapsed > budget {
		t.Fatalf("took %v, which is past the %v a run is allowed", elapsed, budget)
	}
}

func TestReadyPageWritesScreenshotAndArtifact(t *testing.T) {
	options := testOptions(t)
	req := request(t, servePage(t, readyPage), 30000, "out.png", true)

	if !screenshot(options, req) {
		t.Fatal("expected the run to succeed")
	}

	image, err := os.ReadFile(req.destFile)
	if err != nil {
		t.Fatalf("reading the screenshot: %v", err)
	}
	if !strings.HasPrefix(string(image), "\x89PNG") {
		t.Errorf("a .png destination did not get a PNG: % x", image[:8])
	}

	artifact, err := os.ReadFile(req.artifactFile)
	if err != nil {
		t.Fatalf("reading the artifact: %v", err)
	}
	if !strings.Contains(string(artifact), `"html":"<p>hello</p>"`) {
		t.Errorf("the artifact is not what the page returned: %s", artifact)
	}
}

// The screenshot is already written and still good at that point, so a missing hook is reported
// rather than discarding the run's output.
func TestPageWithoutACaptureHookStillWritesTheScreenshot(t *testing.T) {
	options := testOptions(t)
	req := request(t, servePage(t, readyPageNoHook), 30000, "out.png", true)

	if screenshot(options, req) {
		t.Error("expected a missing capture hook to be reported as failure")
	}

	if _, err := os.Stat(req.destFile); err != nil {
		t.Errorf("the screenshot should still have been written: %v", err)
	}
	if _, err := os.Stat(req.artifactFile); err == nil {
		t.Error("no artifact should have been written")
	}
}

// The bug this guards: the format used to follow -quality, so any quality below 100 silently
// produced a JPEG that callers stored and served as a PNG.
func TestDestinationExtensionChoosesTheFormat(t *testing.T) {
	options := testOptions(t)
	req := request(t, servePage(t, readyPage), 30000, "out.jpg", false)

	if !screenshot(options, req) {
		t.Fatal("expected the run to succeed")
	}

	image, err := os.ReadFile(req.destFile)
	if err != nil {
		t.Fatalf("reading the screenshot: %v", err)
	}
	if !strings.HasPrefix(string(image), "\xff\xd8\xff") {
		t.Errorf("a .jpg destination did not get a JPEG: % x", image[:8])
	}
}
