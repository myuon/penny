package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"github.com/myuon/penny/renderer"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	var outputFile string

	rootCmd := &cobra.Command{
		Use:     "penny <input.html or URL>",
		Short:   "penny - a simple HTML renderer",
		Long:    `penny is a command line tool that renders HTML files or URLs to PNG images.`,
		Args:    cobra.ExactArgs(1),
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			var htmlContent string
			var baseURL *url.URL
			var baseDir string

			// Check if input is URL
			if isURL(input) {
				fmt.Printf("Fetching: %s\n", input)
				content, err := fetchURL(input)
				if err != nil {
					return fmt.Errorf("failed to fetch URL: %w", err)
				}
				htmlContent = content
				baseURL, _ = url.Parse(input)
			} else {
				// Read local file
				data, err := os.ReadFile(input)
				if err != nil {
					return fmt.Errorf("failed to read file: %w", err)
				}
				htmlContent = string(data)
				baseDir = filepath.Dir(input)
			}

			// Parse HTML
			document, err := dom.ParseString(htmlContent)
			if err != nil {
				return fmt.Errorf("failed to parse HTML: %w", err)
			}

			// Find and load CSS files from <link> tags
			var stylesheet *css.Stylesheet
			if baseURL != nil {
				stylesheet = loadStylesheetsFromURL(document, baseURL)
			} else {
				stylesheet = loadStylesheetsFromDir(document, baseDir)
			}

			// Ensure output directory exists
			outputDir := filepath.Dir(outputFile)
			if outputDir != "." {
				if err := os.MkdirAll(outputDir, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
			}

			// Render to PNG
			r := renderer.New(800, 600)
			if err := r.Render(document, stylesheet, outputFile); err != nil {
				return fmt.Errorf("failed to render: %w", err)
			}

			fmt.Printf("Rendered to %s\n", outputFile)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "output.png", "output file path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
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

func loadStylesheetsFromDir(d *dom.DOM, baseDir string) *css.Stylesheet {
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
						fmt.Printf("Loaded CSS: %s\n", cssPath)
					}
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
						fmt.Printf("Loaded CSS: %s\n", cssURL)
					}
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
