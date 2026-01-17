package reftest

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"github.com/myuon/penny/layout"
	"github.com/myuon/penny/paint"
	"github.com/playwright-community/playwright-go"
)

const (
	viewportWidth  = 800
	viewportHeight = 600
)

type ReftestResult struct {
	Name          string
	DiffPercent   float64
	ChromeImage   *image.RGBA
	PennyImage    *image.RGBA
	DiffImage     *image.RGBA
	CombinedImage *image.RGBA
}

func TestReftest(t *testing.T) {
	// Find test files
	testDataDir := "testdata"
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata dir: %v", err)
	}

	var htmlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".html" {
			htmlFiles = append(htmlFiles, filepath.Join(testDataDir, entry.Name()))
		}
	}

	if len(htmlFiles) == 0 {
		t.Skip("no HTML test files found in testdata/")
	}

	// Start local HTTP server
	server := startTestServer(testDataDir)
	defer server.Close()

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Run tests
	for _, htmlFile := range htmlFiles {
		testName := filepath.Base(htmlFile)
		testName = testName[:len(testName)-5] // remove .html

		t.Run(testName, func(t *testing.T) {
			result, err := runReftest(browser, server.Addr, htmlFile, testName)
			if err != nil {
				t.Fatalf("reftest failed: %v", err)
			}

			// Save combined image
			outputPath := filepath.Join(outputDir, testName+"_diff.png")
			if err := savePNG(result.CombinedImage, outputPath); err != nil {
				t.Errorf("failed to save diff image: %v", err)
			}

			t.Logf("Diff: %.2f%% - Output: %s", result.DiffPercent, outputPath)

			// Optionally fail if diff is too high
			// if result.DiffPercent > 5.0 {
			// 	t.Errorf("diff too high: %.2f%%", result.DiffPercent)
			// }
		})
	}
}

func startTestServer(dir string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(dir)))

	server := &http.Server{
		Addr:    "127.0.0.1:9876",
		Handler: mux,
	}

	go server.ListenAndServe()
	return server
}

func runReftest(browser playwright.Browser, serverAddr, htmlFile, testName string) (*ReftestResult, error) {
	// Get Chrome screenshot
	chromeImg, err := captureChrome(browser, serverAddr, filepath.Base(htmlFile))
	if err != nil {
		return nil, fmt.Errorf("chrome capture failed: %w", err)
	}

	// Get Penny rendering
	pennyImg, err := capturePenny(htmlFile)
	if err != nil {
		return nil, fmt.Errorf("penny render failed: %w", err)
	}

	// Compare images
	diffImg, diffPercent := compareImages(chromeImg, pennyImg)

	// Create combined image (Chrome | Penny | Diff)
	combinedImg := createCombinedImage(chromeImg, pennyImg, diffImg)

	return &ReftestResult{
		Name:          testName,
		DiffPercent:   diffPercent,
		ChromeImage:   chromeImg,
		PennyImage:    pennyImg,
		DiffImage:     diffImg,
		CombinedImage: combinedImg,
	}, nil
}

func captureChrome(browser playwright.Browser, serverAddr, htmlFileName string) (*image.RGBA, error) {
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  viewportWidth,
			Height: viewportHeight,
		},
	})
	if err != nil {
		return nil, err
	}
	defer page.Close()

	url := fmt.Sprintf("http://%s/%s", serverAddr, htmlFileName)
	if _, err := page.Goto(url); err != nil {
		return nil, err
	}

	// Wait for page to load
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, err
	}

	// Take screenshot
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		return nil, err
	}

	// Decode PNG
	return decodePNG(screenshot)
}

func capturePenny(htmlFile string) (*image.RGBA, error) {
	// Read HTML file
	htmlContent, err := os.ReadFile(htmlFile)
	if err != nil {
		return nil, err
	}

	// Parse HTML
	document, err := dom.ParseString(string(htmlContent))
	if err != nil {
		return nil, err
	}

	// Load CSS
	baseDir := filepath.Dir(htmlFile)
	stylesheet := loadStylesheets(document, baseDir)

	// Build layout tree
	layoutTree := layout.BuildLayoutTree(document, stylesheet)

	// Compute layout
	layout.ComputeLayout(layoutTree, viewportWidth, viewportHeight)

	// Paint
	paintList := paint.NewPaintList()
	paint.PaintBackground(paintList, viewportWidth, viewportHeight, css.ColorWhite)
	ops := paint.Paint(layoutTree)
	paintList.Ops = append(paintList.Ops, ops.Ops...)

	// Rasterize
	img := paint.Rasterize(paintList, viewportWidth, viewportHeight)
	return img, nil
}

func loadStylesheets(d *dom.DOM, baseDir string) *css.Stylesheet {
	var allRules []css.Rule

	var walk func(nodeID dom.NodeID)
	walk = func(nodeID dom.NodeID) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "link" {
			rel, hasRel := node.Attr["rel"]
			href, hasHref := node.Attr["href"]
			if hasRel && rel == "stylesheet" && hasHref {
				cssPath := filepath.Join(baseDir, href)
				if data, err := os.ReadFile(cssPath); err == nil {
					if sheet, err := css.Parse(string(data)); err == nil {
						allRules = append(allRules, sheet.Rules...)
					}
				}
			}
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "style" {
			cssText := extractTextContent(d, nodeID)
			if cssText != "" {
				if sheet, err := css.Parse(cssText); err == nil {
					allRules = append(allRules, sheet.Rules...)
				}
			}
		}

		for _, childID := range node.Children {
			walk(childID)
		}
	}

	walk(d.Root)

	if len(allRules) == 0 {
		return nil
	}

	return &css.Stylesheet{Rules: allRules}
}

func extractTextContent(d *dom.DOM, nodeID dom.NodeID) string {
	var text string
	var walk func(id dom.NodeID)
	walk = func(id dom.NodeID) {
		node := d.GetNode(id)
		if node == nil {
			return
		}
		if node.Type == dom.NodeTypeText {
			text += node.Text
		}
		for _, childID := range node.Children {
			walk(childID)
		}
	}
	walk(nodeID)
	return text
}

func compareImages(img1, img2 *image.RGBA) (*image.RGBA, float64) {
	bounds := img1.Bounds()
	diffImg := image.NewRGBA(bounds)

	totalPixels := bounds.Dx() * bounds.Dy()
	diffPixels := 0

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)

			if colorsEqual(c1, c2) {
				// Same pixel - show dimmed version
				diffImg.SetRGBA(x, y, color.RGBA{
					R: c1.R / 3,
					G: c1.G / 3,
					B: c1.B / 3,
					A: 255,
				})
			} else {
				// Different pixel - show in red
				diffImg.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
				diffPixels++
			}
		}
	}

	diffPercent := float64(diffPixels) / float64(totalPixels) * 100
	return diffImg, diffPercent
}

func colorsEqual(c1, c2 color.RGBA) bool {
	// Allow small tolerance for anti-aliasing differences
	const tolerance = 5
	return abs(int(c1.R)-int(c2.R)) <= tolerance &&
		abs(int(c1.G)-int(c2.G)) <= tolerance &&
		abs(int(c1.B)-int(c2.B)) <= tolerance
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func createCombinedImage(chrome, penny, diff *image.RGBA) *image.RGBA {
	bounds := chrome.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create combined image: Chrome | Penny | Diff
	combined := image.NewRGBA(image.Rect(0, 0, width*3, height+30))

	// Fill with gray background
	draw.Draw(combined, combined.Bounds(), &image.Uniform{color.RGBA{40, 40, 40, 255}}, image.Point{}, draw.Src)

	// Draw Chrome image
	draw.Draw(combined, image.Rect(0, 30, width, height+30), chrome, bounds.Min, draw.Src)

	// Draw Penny image
	draw.Draw(combined, image.Rect(width, 30, width*2, height+30), penny, bounds.Min, draw.Src)

	// Draw Diff image
	draw.Draw(combined, image.Rect(width*2, 30, width*3, height+30), diff, bounds.Min, draw.Src)

	return combined
}

func decodePNG(data []byte) (*image.RGBA, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba, nil
}

func savePNG(img *image.RGBA, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// TestReftestURLs runs reftests against URLs listed in urls.txt
func TestReftestURLs(t *testing.T) {
	urlsFile := "testdata/urls.txt"
	urls, err := readURLsFile(urlsFile)
	if err != nil {
		t.Skipf("no urls.txt found: %v", err)
	}

	if len(urls) == 0 {
		t.Skip("no URLs in urls.txt")
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch()
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create output directory
	outputDir := "output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Run tests for each URL
	for _, testURL := range urls {
		testName := urlToTestName(testURL)

		t.Run(testName, func(t *testing.T) {
			result, err := runReftestURL(browser, testURL, testName)
			if err != nil {
				t.Fatalf("reftest failed: %v", err)
			}

			// Save combined image
			outputPath := filepath.Join(outputDir, testName+"_diff.png")
			if err := savePNG(result.CombinedImage, outputPath); err != nil {
				t.Errorf("failed to save diff image: %v", err)
			}

			t.Logf("Diff: %.2f%% - Output: %s", result.DiffPercent, outputPath)
		})
	}
}

func readURLsFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}

	return urls, scanner.Err()
}

func urlToTestName(testURL string) string {
	// Create a short name from URL using MD5 hash
	parsed, err := url.Parse(testURL)
	if err != nil {
		return fmt.Sprintf("%x", md5.Sum([]byte(testURL)))[:8]
	}

	// Use host + path, sanitized
	name := parsed.Host + parsed.Path
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, ":", "_")

	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

func runReftestURL(browser playwright.Browser, testURL, testName string) (*ReftestResult, error) {
	// Get Chrome screenshot
	chromeImg, err := captureChromeURL(browser, testURL)
	if err != nil {
		return nil, fmt.Errorf("chrome capture failed: %w", err)
	}

	// Get Penny rendering
	pennyImg, err := capturePennyURL(testURL)
	if err != nil {
		return nil, fmt.Errorf("penny render failed: %w", err)
	}

	// Compare images
	diffImg, diffPercent := compareImages(chromeImg, pennyImg)

	// Create combined image (Chrome | Penny | Diff)
	combinedImg := createCombinedImage(chromeImg, pennyImg, diffImg)

	return &ReftestResult{
		Name:          testName,
		DiffPercent:   diffPercent,
		ChromeImage:   chromeImg,
		PennyImage:    pennyImg,
		DiffImage:     diffImg,
		CombinedImage: combinedImg,
	}, nil
}

func captureChromeURL(browser playwright.Browser, testURL string) (*image.RGBA, error) {
	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{
			Width:  viewportWidth,
			Height: viewportHeight,
		},
	})
	if err != nil {
		return nil, err
	}
	defer page.Close()

	if _, err := page.Goto(testURL); err != nil {
		return nil, err
	}

	// Wait for page to load
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, err
	}

	// Take screenshot
	screenshot, err := page.Screenshot(playwright.PageScreenshotOptions{
		Type: playwright.ScreenshotTypePng,
	})
	if err != nil {
		return nil, err
	}

	return decodePNG(screenshot)
}

func capturePennyURL(testURL string) (*image.RGBA, error) {
	// Fetch HTML content
	htmlContent, err := fetchURL(testURL)
	if err != nil {
		return nil, err
	}

	// Parse HTML
	document, err := dom.ParseString(htmlContent)
	if err != nil {
		return nil, err
	}

	// Parse base URL for CSS loading
	baseURL, err := url.Parse(testURL)
	if err != nil {
		return nil, err
	}

	// Load CSS from URL
	stylesheet := loadStylesheetsFromURL(document, baseURL)

	// Build layout tree
	layoutTree := layout.BuildLayoutTree(document, stylesheet)

	// Compute layout
	layout.ComputeLayout(layoutTree, viewportWidth, viewportHeight)

	// Paint
	paintList := paint.NewPaintList()
	paint.PaintBackground(paintList, viewportWidth, viewportHeight, css.ColorWhite)
	ops := paint.Paint(layoutTree)
	paintList.Ops = append(paintList.Ops, ops.Ops...)

	// Rasterize
	img := paint.Rasterize(paintList, viewportWidth, viewportHeight)
	return img, nil
}

func fetchURL(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func loadStylesheetsFromURL(d *dom.DOM, baseURL *url.URL) *css.Stylesheet {
	var allRules []css.Rule

	var walk func(nodeID dom.NodeID)
	walk = func(nodeID dom.NodeID) {
		node := d.GetNode(nodeID)
		if node == nil {
			return
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "link" {
			rel, hasRel := node.Attr["rel"]
			href, hasHref := node.Attr["href"]
			if hasRel && rel == "stylesheet" && hasHref {
				cssURL := resolveURL(baseURL, href)
				if content, err := fetchURL(cssURL); err == nil {
					if sheet, err := css.Parse(content); err == nil {
						allRules = append(allRules, sheet.Rules...)
					}
				}
			}
		}

		if node.Type == dom.NodeTypeElement && node.Tag == "style" {
			cssText := extractTextContent(d, nodeID)
			if cssText != "" {
				if sheet, err := css.Parse(cssText); err == nil {
					allRules = append(allRules, sheet.Rules...)
				}
			}
		}

		for _, childID := range node.Children {
			walk(childID)
		}
	}

	walk(d.Root)

	if len(allRules) == 0 {
		return nil
	}

	return &css.Stylesheet{Rules: allRules}
}

func resolveURL(base *url.URL, ref string) string {
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(refURL).String()
}
