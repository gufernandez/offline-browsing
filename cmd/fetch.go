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

type weblink struct {
	baseUrl  string
	url      string
	filename string
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <url>",
	Short: "Download your webpage for offline use.",
	Long: `Use this application to download all webpages that you like. 
	
	The informed URLs will be downloaded to the current folder as html files.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please specify at least one URL")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		showMetadata, _ := cmd.Flags().GetBool("metadata")
		root(args, showMetadata)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := fetchCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Flags definition
func init() {
	fetchCmd.Flags().BoolP("metadata", "m", false, "Show metadata of existing file. If non existent, will be downloaded.")
}

// Root function to handle args
func root(urls []string, showMetadata bool) {
	for _, url := range urls {
		fetchLink := formatLink(url)
		if showMetadata {
			if _, err := os.Stat(fetchLink.filename); os.IsNotExist(err) {
				fmt.Println("-- Downloading Now")
				fetch(fetchLink)
				printMetadata(fetchLink.filename)
			} else {
				fmt.Println("-- Already Downloaded")
				printMetadata(fetchLink.filename)
			}
		} else {
			fetch(fetchLink)
		}
		fmt.Println()
	}
}

func fetch(fetchLink weblink) {
	resp, err := http.Get(fetchLink.url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	f, err := os.Create(fetchLink.filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}

	setMetadata(fetchLink)
}

func printMetadata(filename string) {
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

func setMetadata(fetchLink weblink) {
	buf := bytes.NewBuffer(nil)
	f, err := os.OpenFile(fetchLink.filename, os.O_APPEND|os.O_RDWR, 0644)
	io.Copy(buf, f)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	s := buf.String()
	dt := time.Now()

	siteMeta := metadataFormat("site", fetchLink.baseUrl)
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

func formatLink(inputUrl string) weblink {
	r, _ := regexp.Compile(`(https?:\/\/)*(.+)`)
	match := r.FindAllStringSubmatch(inputUrl, -1)

	protocol := match[0][1]
	inputBaseUrl := match[0][2]

	if inputBaseUrl == "" {
		panic("Invalid URL: " + inputUrl)
	}

	if protocol == "" {
		protocol = "https://"
	}

	filePath := strings.ReplaceAll(inputBaseUrl, "/", ".") + ".html"

	return weblink{baseUrl: inputBaseUrl, url: protocol + inputBaseUrl, filename: filePath}
}
