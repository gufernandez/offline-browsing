package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download your webpage for offline use.",
	Long: `Use this application to download all webpages that you like. 
	
	The informed URLs will be downloaded to the current folder as html files.`,
	Run: func(cmd *cobra.Command, args []string) {
		showMetadata, _ := cmd.Flags().GetBool("metadata")
		root(args, showMetadata)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Flags definition
func init() {
	rootCmd.Flags().BoolP("metadata", "m", false, "Show metadata of existing file. If non existent, will be downloaded.")
}

// Root function to handle args
func root(urls []string, showMetadata bool) {
	for _, url := range urls {
		if showMetadata {
			filename := urlFilename(url)
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				fetch(url)
				getMetadata(filename)
			} else {
				getMetadata(filename)
			}
		} else {
			fetch(url)
		}
	}
}

func getMetadata(filename string) {
	buf := bytes.NewBuffer(nil)
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	io.Copy(buf, f)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s := buf.String()

	r, _ := regexp.Compile("<meta name=\"cmd-(.+)\" content=\"(.+)\">")
	match := r.FindAllStringSubmatch(s, -1)
	for _, tag := range match {
		fmt.Println(tag[1] + ": " + tag[2])
	}
}

func fetch(url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	filename := urlFilename(url)
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	setMetadata(url)
}

func setMetadata(url string) {
	buf := bytes.NewBuffer(nil)
	f, err := os.OpenFile(urlFilename(url), os.O_APPEND|os.O_RDWR, 0644)
	io.Copy(buf, f)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s := buf.String()
	dt := time.Now()

	siteMeta := metadataFormat("site", url)
	linksMeta := metadataFormat("num_links", getLinksQuantity(s))
	imagesMeta := metadataFormat("images", getImagesQuantity(s))
	dateMeta := metadataFormat("last_fetch", dt.Format(time.RFC1123))
	toAppend := siteMeta + linksMeta + imagesMeta + dateMeta
	if _, err := f.WriteString(toAppend); err != nil {
		panic(err)
	}
}

func getLinksQuantity(fileContent string) string {
	r, _ := regexp.Compile("(href=\"https)")
	match := r.FindAllString(fileContent, -1)
	return fmt.Sprintf("%d", len(match))
}

func getImagesQuantity(fileContent string) string {
	r, _ := regexp.Compile("(<img)")
	match := r.FindAllString(fileContent, -1)
	return fmt.Sprintf("%d", len(match))
}

func metadataFormat(name string, content string) string {
	return "\n<meta name=\"cmd-" + name + "\" content=\"" + content + "\">"
}

func urlFilename(url string) string {
	return strings.Split(url, "//")[1] + ".html"
}
