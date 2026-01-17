package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/myuon/penny/css"
	"github.com/myuon/penny/dom"
	"github.com/myuon/penny/renderer"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	var outputFile string

	rootCmd := &cobra.Command{
		Use:     "penny <input.html>",
		Short:   "penny - a simple HTML renderer",
		Long:    `penny is a command line tool that renders HTML files to PNG images.`,
		Args:    cobra.ExactArgs(1),
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			inputDir := filepath.Dir(inputFile)

			// Read input file
			file, err := os.Open(inputFile)
			if err != nil {
				return fmt.Errorf("failed to open input file: %w", err)
			}
			defer file.Close()

			// Parse HTML
			document, err := dom.Parse(file)
			if err != nil {
				return fmt.Errorf("failed to parse HTML: %w", err)
			}

			// Find and load CSS files from <link> tags
			stylesheet := loadStylesheets(document, inputDir)

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

func loadStylesheets(d *dom.DOM, baseDir string) *css.Stylesheet {
	var allRules []css.Rule

	// Walk DOM to find <link rel="stylesheet"> tags
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
