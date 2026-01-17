package reftest

import (
	"encoding/json"
	"fmt"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playwright-community/playwright-go"
)

const wptRoot = "../wpt"

// WPTTestResult holds the result of a single WPT test
type WPTTestResult struct {
	Name        string  `json:"name"`
	URL         string  `json:"url"`
	DiffPercent float64 `json:"diff_percent"`
	Status      string  `json:"status"` // "pass", "fail", "error"
	Error       string  `json:"error,omitempty"`
}

// WPTSuiteResult holds the results of a WPT test suite
type WPTSuiteResult struct {
	Suite      string          `json:"suite"`
	Total      int             `json:"total"`
	Passed     int             `json:"passed"`
	Failed     int             `json:"failed"`
	Errors     int             `json:"errors"`
	Results    []WPTTestResult `json:"results"`
	Threshold  float64         `json:"threshold"`
}

// TestWPTFlexbox runs WPT css-flexbox tests
func TestWPTFlexbox(t *testing.T) {
	runWPTSuite(t, "css/css-flexbox", 10.0) // 10% threshold
}

// runWPTSuite runs all HTML tests in a WPT suite directory
func runWPTSuite(t *testing.T, suite string, threshold float64) {
	suiteDir := filepath.Join(wptRoot, suite)

	// Check if WPT is available
	if _, err := os.Stat(suiteDir); os.IsNotExist(err) {
		t.Skipf("WPT suite not found: %s (run 'git submodule update --init')", suiteDir)
	}

	// Find all HTML test files
	var testFiles []string
	err := filepath.Walk(suiteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm")) {
			// Skip reference files (used for WPT reftests)
			if strings.Contains(path, "-ref.") || strings.Contains(path, "-ref-") {
				return nil
			}
			// Skip support files
			if strings.Contains(path, "/support/") {
				return nil
			}
			testFiles = append(testFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk suite directory: %v", err)
	}

	if len(testFiles) == 0 {
		t.Skip("no test files found")
	}

	t.Logf("Found %d test files in %s", len(testFiles), suite)

	// Randomly select tests (full suite takes too long)
	maxTests := 50
	if len(testFiles) > maxTests {
		t.Logf("Randomly selecting %d tests from %d", maxTests, len(testFiles))
		rand.Shuffle(len(testFiles), func(i, j int) {
			testFiles[i], testFiles[j] = testFiles[j], testFiles[i]
		})
		testFiles = testFiles[:maxTests]
	}

	// Start HTTP server for WPT files
	server := startTestServer(wptRoot)
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
	outputDir := filepath.Join("output", "wpt", suite)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Run tests
	suiteResult := &WPTSuiteResult{
		Suite:     suite,
		Threshold: threshold,
	}

	for _, testFile := range testFiles {
		relPath, _ := filepath.Rel(wptRoot, testFile)
		testName := strings.ReplaceAll(relPath, "/", "_")
		testName = strings.TrimSuffix(testName, ".html")
		testName = strings.TrimSuffix(testName, ".htm")

		t.Run(testName, func(t *testing.T) {
			result := runWPTTest(t, browser, server.Addr, testFile, relPath, outputDir, threshold)
			suiteResult.Results = append(suiteResult.Results, result)
			suiteResult.Total++

			switch result.Status {
			case "pass":
				suiteResult.Passed++
			case "fail":
				suiteResult.Failed++
			case "error":
				suiteResult.Errors++
			}
		})
	}

	// Save summary
	summaryPath := filepath.Join(outputDir, "summary.json")
	if data, err := json.MarshalIndent(suiteResult, "", "  "); err == nil {
		os.WriteFile(summaryPath, data, 0644)
	}

	t.Logf("WPT Suite %s: %d/%d passed (%.1f%%), %d errors",
		suite, suiteResult.Passed, suiteResult.Total,
		float64(suiteResult.Passed)/float64(suiteResult.Total)*100,
		suiteResult.Errors)
}

func runWPTTest(t *testing.T, browser playwright.Browser, serverAddr, testFile, relPath, outputDir string, threshold float64) WPTTestResult {
	testURL := fmt.Sprintf("http://%s/%s", serverAddr, relPath)

	result := WPTTestResult{
		Name: relPath,
		URL:  testURL,
	}

	// Get Chrome screenshot
	chromeImg, err := captureChromeURL(browser, testURL)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("chrome capture failed: %v", err)
		t.Logf("ERROR: %s", result.Error)
		return result
	}

	// Get Penny rendering
	pennyImg, err := capturePennyFile(testFile)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("penny render failed: %v", err)
		t.Logf("ERROR: %s", result.Error)
		return result
	}

	// Compare images
	diffImg, diffPercent := compareImages(chromeImg, pennyImg)
	result.DiffPercent = diffPercent

	// Determine pass/fail
	if diffPercent <= threshold {
		result.Status = "pass"
		t.Logf("PASS: %.2f%% diff", diffPercent)
	} else {
		result.Status = "fail"
		t.Logf("FAIL: %.2f%% diff (threshold: %.2f%%)", diffPercent, threshold)
	}

	// Save diff image for all tests
	combinedImg := createCombinedImage(chromeImg, pennyImg, diffImg)
	testName := strings.ReplaceAll(relPath, "/", "_")
	outputPath := filepath.Join(outputDir, testName+"_diff.png")
	savePNG(combinedImg, outputPath)

	return result
}

func capturePennyFile(htmlFile string) (*image.RGBA, error) {
	return capturePenny(htmlFile)
}
