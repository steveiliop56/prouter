package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

var (
	Version = "dev"
)

func main() {
	// Configure zerolog
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create flagset
	flagSet := pflag.NewFlagSet("prouter", pflag.ExitOnError)

	// Define command line flags
	flagSet.String("serve", "", "Path to serve static files from")
	flagSet.String("address", "", "Address to bind the server to (default all interfaces)")
	flagSet.String("port", "8080", "Port to bind the server to (default 8080)")

	// Parse command line flags
	flagSet.Parse(os.Args[1:])

	// Log the version
	log.Info().Str("version", Version).Msg("Starting prouter")

	// Get serve path
	servePath, err := flagSet.GetString("serve")

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get serve path")
	}

	// Get address
	address, err := flagSet.GetString("address")

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get address")
	}

	// Get port
	port, err := flagSet.GetString("port")

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get port")
	}

	// Ensure the serve path exists
	if _, err := os.Stat(servePath); os.IsNotExist(err) {
		log.Fatal().Err(err).Str("servePath", servePath).Msg("Serve path does not exist")
	}

	log.Info().Str("servePath", servePath).Msg("Using serve path")

	// Create the HTML template
	htmlTemplate, err := template.New("index").
		Parse(`<!DOCTYPE html>
		<html>
			<head>
				<title>{{.Title}}</title>
				<meta name="viewport" content="width=device-width, initial-scale=1.0" />
				<style>
					body {
						background-color: #0d1117;
						color: #f0f6fc;
						font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans",
							Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
						display: flex;
						flex-direction: column;
						padding: 0.5rem;
					}

					h1,
					h2 {
						border-bottom: 1px solid #3d444db3;
						padding-bottom: 0.3rem;
					}
				</style>
			</head>
			<body>
				{{.Content}}
			</body>
		</html>
		`)

	// Check for template creation error
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create HTML template")
	}

	// Create handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		log.Info().Str("method", r.Method).Str("url", r.URL.String()).Str("host", r.Host).Msg("Received request")

		// Get subdomain
		subdomain := strings.Split(r.Host, ".")[0]

		// Create the file path
		filePath := path.Join(servePath, subdomain)

		// Check if the path exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Log the error
			log.Error().Err(err).Str("filePath", filePath).Msg("Path does not exist")

			// Return error
			http.Error(w, fmt.Sprintf("Site not found: %s", subdomain), http.StatusNotFound)
			return
		}

		// Get directory contents
		files, err := os.ReadDir(filePath)

		if err != nil {
			// Log the error
			log.Error().Err(err).Str("filePath", filePath).Msg("Failed to read directory")

			// Return error
			http.Error(w, fmt.Sprintf("Failed to read directory: %s", err), http.StatusInternalServerError)
			return
		}

		// If no files, return 404
		if len(files) == 0 {
			// Log the error
			log.Error().Str("filePath", filePath).Msg("No files found in directory")

			// Return error
			http.Error(w, fmt.Sprintf("No files found in site: %s", subdomain), http.StatusNotFound)
			return
		}

		// Walk directory
		err = filepath.WalkDir(filePath, func(pSrc string, d fs.DirEntry, err error) error {
			if err != nil {
				// Log the error
				log.Error().Err(err).Str("path", pSrc).Msg("Failed to walk directory")

				// Return error
				http.Error(w, fmt.Sprintf("Failed to walk directory: %s", err), http.StatusInternalServerError)

				// Return the error to stop walking
				return err
			}

			// Check if the entry is a directory
			if d.IsDir() {
				return nil // Skip directories
			}

			// Remove the "public/" prefix and the file extension from the path
			p := strings.TrimPrefix(pSrc, "public/")

			// Remove subdomain from the path
			p = strings.TrimPrefix(p, subdomain+"/")

			// Remove the leading slash if it exists
			url := strings.TrimPrefix(r.URL.Path, "/")

			// If the URL ends with a slash, remove it
			url = strings.TrimSuffix(url, "/")

			// Check if the file matches the request
			if strings.TrimSuffix(p, path.Ext(p)) == url || p == url {
				log.Info().Str("fileName", d.Name()).Msg("Found requested file")

				// Check if the file is a markdown file
				if strings.HasSuffix(p, ".md") {
					// Render the markdown file
					renderMarkdown(pSrc, w, htmlTemplate)

					// Return
					return errors.New("rendered")
				} else {
					// Serve the file directly
					http.ServeFile(w, r, pSrc)

					// Return
					return errors.New("rendered")
				}
			}

			return nil
		})

		// If there was an error it means we found the file and rendered it
		if err != nil && err.Error() == "rendered" {
			return
		}

		// Check if there is an index.md file
		indexMdFile := path.Join(filePath, "index.md")

		if _, err := os.Stat(indexMdFile); err == nil {
			// Render the index.md file
			renderMarkdown(indexMdFile, w, htmlTemplate)

			// Return
			return
		}

		// Serve whatever file is in the directory
		server := http.FileServer(http.Dir(filePath))

		// Serve the files
		server.ServeHTTP(w, r)
	}

	// Start the HTTP server
	listen := address + ":" + port

	log.Info().Str("listen", listen).Msg("Starting server")

	log.Fatal().Err(http.ListenAndServe(listen, http.HandlerFunc(handler))).Msg("Server failed")
}

func renderMarkdown(filePath string, w http.ResponseWriter, t *template.Template) {
	// Read the markdown file
	data, err := os.ReadFile(filePath)

	if err != nil {
		// Log the error
		log.Error().Err(err).Str("filePath", filePath).Msg("Failed to read markdown file")

		// Return error
		http.Error(w, fmt.Sprintf("Failed to read markdown file: %s", err), http.StatusInternalServerError)
		return
	}

	// Create markdown parser
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	md := parser.NewWithExtensions(extensions)

	// Create HTML renderer with common flags
	flags := html.CommonFlags | html.HrefTargetBlank
	options := html.RendererOptions{Flags: flags}
	renderer := html.NewRenderer(options)

	// Parse the markdown file
	document := md.Parse(data)

	// Render the markdown to HTML
	html := markdown.Render(document, renderer)

	// Create a template variables map
	vars := map[string]interface{}{
		"Title":   strings.TrimSuffix(path.Base(filePath), path.Ext(filePath)),
		"Content": template.HTML(html),
	}

	// Execute the template with the variables
	err = t.Execute(w, vars)

	// Check for write error
	if err != nil {
		// Log the error
		log.Error().Err(err).Str("filePath", filePath).Msg("Failed to write HTML response")

		// Return error
		http.Error(w, fmt.Sprintf("Failed to write HTML response: %s", err), http.StatusInternalServerError)
	}
}
