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
		fullDownload, _ := cmd.Flags().GetBool("full-download")
		root(args, showMetadata, fullDownload)
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
	fetchCmd.Flags().BoolP("full-download", "f", false, "Download all images from the HTML for total offline use")
}

// Root function to handle args
func root(urls []string, showMetadata bool, fullDownload bool) {
	for _, url := range urls {
		fetchLink := formatLink(url)
		if showMetadata {
			if _, err := os.Stat(fetchLink.filename); os.IsNotExist(err) {
				fmt.Println("-- Downloading Now")
				fetch(fetchLink, fullDownload)
				printFileMetadata(fetchLink.filename)
			} else {
				fmt.Println("-- Already Downloaded")
				printFileMetadata(fetchLink.filename)
			}
		} else {
			fetch(fetchLink, fullDownload)
		}
		fmt.Println()
	}
}

func fetch(fetchLink weblink, downloadAll bool) {
	saveLinkContentToFile(fetchLink.url, fetchLink.filename)
	setFileMetadataAndImages(fetchLink, downloadAll)
}

func saveLinkContentToFile(url string, filepath string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}
}

func printFileMetadata(filename string) {
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

func setFileMetadataAndImages(fetchLink weblink, fetchImages bool) {
	buf := bytes.NewBuffer(nil)
	f, err := os.OpenFile(fetchLink.filename, os.O_RDWR|os.O_APPEND, 0644)
	io.Copy(buf, f)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fileContent := buf.String()

	imagesCount, imagesLinks := getImagesLinks(fileContent)

	toAppend := metadataFormat("site", fetchLink.baseUrl) + metadataFormat("num_links", getLinksQuantity(fileContent)) + metadataFormat("images", imagesCount) + metadataFormat("last_fetch", time.Now().Format(time.RFC1123))
	if _, err := f.WriteString(toAppend); err != nil {
		panic(err)
	}

	if fetchImages {
		updatedContent := downloadImages(imagesLinks, fetchLink, fileContent+toAppend)
		updateFileContent(fetchLink.filename, updatedContent)
	}

}

func updateFileContent(filepath string, content string) {
	f, err := os.Create(filepath)
	if err != nil {
		panic(err)
	}
	_, err = fmt.Fprint(f, content)
	if err != nil {
		panic(err)
	}
}

func downloadImages(linkList []string, pageLink weblink, fileContent string) string {
	srcR, _ := regexp.Compile(`[^\/]\/([^\/]+\.[a-z]+)`)
	os.Mkdir(pageLink.baseUrl, os.ModePerm)
	for _, img := range linkList {
		name := srcR.FindAllStringSubmatch(img, -1)
		newSrc := pageLink.baseUrl + "/" + name[0][1]

		imgUrl := img
		if strings.HasPrefix(img, "/") {
			imgUrl = pageLink.url + img
		}

		saveLinkContentToFile(imgUrl, newSrc)
		fileContent = strings.ReplaceAll(fileContent, img, newSrc)
	}
	return fileContent
}

func getLinksQuantity(fileContent string) string {
	r, _ := regexp.Compile("(href=\"http)")
	match := r.FindAllString(fileContent, -1)
	return fmt.Sprintf("%d", len(match))
}

func getImagesLinks(fileContent string) (string, []string) {
	r, _ := regexp.Compile(`<img[^>]*src="([^"]*)"`)
	match := r.FindAllStringSubmatch(fileContent, -1)
	linkList := make([]string, len(match))
	for i := range match {
		linkList[i] = match[i][1]
	}
	return fmt.Sprintf("%d", len(linkList)), linkList
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
