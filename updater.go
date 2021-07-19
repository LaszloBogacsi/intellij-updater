package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/dustin/go-humanize"
)

type Build struct {
	XMLName xml.Name `xml:"build"`
	Version string   `xml:"version,attr"`
}

type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Build   []Build  `xml:"build"`
	Id      string   `xml:"id,attr"`
}

type Code struct {
	XMLName xml.Name `xml:"code"`
}

type Product struct {
	XMLName  xml.Name  `xml:"product"`
	Name     string    `xml:"name,attr"`
	Code     []Code    `xml:"code"`
	Channels []Channel `xml:"channel"`
}

type Products struct {
	XMLName xml.Name  `xml:"products"`
	Product []Product `xml:"product"`
}

func main() {
	url := "https://www.jetbrains.com/updates/updates.xml"
	var products Products
	responseBody, _ := getXML(url)
	err := parseXML(responseBody, &products)
	if err != nil {
		return
	}
	// get product by name,
	const productName = "IntelliJ IDEA"
	intelliJ := filterProduct(products.Product, func(product Product) bool {
		return product.Name == productName
	})

	// get correct channel
	const channelId = "IC-IU-RELEASE-licensing-RELEASE"
	intelliJReleaseChannel := filterChannel(intelliJ[0].Channels, func(channel Channel) bool {
		return channel.Id == channelId
	})

	// get latest build
	latestBuildVersion := intelliJReleaseChannel[0].Build[0].Version

	// get version from latest build
	fmt.Println("Latest version found: ", latestBuildVersion)

	// assemble url
	filename := "ideaIU-" + latestBuildVersion + ".dmg"
	downloadUrl := "https://download.jetbrains.com/idea/" + filename

	// download from url if file not exist
	if _, err := os.Stat(filename); err == nil {
		// path/to/whatever exists
		fmt.Println("File already exists: ", filename)
		fmt.Print("Would you like download it anyway? [Y/n]")
		reader := bufio.NewReader(os.Stdin)
		char, _, readErr := reader.ReadRune()
		if readErr != nil {
			fmt.Println(readErr)
		}

		switch char {
		case 'Y':
			fmt.Println("Downloading...")
			downloadFile(err, filename, downloadUrl)
			break
		case 'n':
			fmt.Println("Downloading cancelled...")
			break
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		downloadFile(err, filename, downloadUrl)
	} else {

	}

	/*
		removing files
	*/
	var filesToRemove = []string{
		"/System/Volumes/Data/Users/%[1]s/Library/Saved Application State/com.jetbrains.intellij.savedState",
		"/System/Volumes/Data/Users/%[1]s/Library/Application Support/JetBrains",
		"/System/Volumes/Data/Users/%[1]s/Library/Logs/JetBrains",
		"/System/Volumes/Data/Users/%[1]s/Library/Caches/JetBrains",
		"/System/Volumes/Data/Users/%[1]s/Library/Caches/com.apple.python/Users/%[1]s/Library/Application Support/JetBrains",
		"/Users/%[1]s/Library/Saved Application State/com.jetbrains.intellij.savedState",
		"/Users/%[1]s/Library/Application Support/JetBrains",
		"/Users/%[1]s/Library/Logs/JetBrains",
		"/Users/%[1]s/Library/Caches/JetBrains",
		"/Users/%[1]s/Library/Caches/com.apple.python/Users/%[1]s/Library/Application Support/JetBrains",
		"/Users/%[1]s/Library/Preferences/jetbrains.idea.*.plist",
		"/Users/%[1]s/Library/Preferences/com.jetbrains.intellij.plist",
	}
	var filesToRemove2 = []string{
		"/System/Volumes/Data/private/var/folders/8q/5llq9zx167d0__x3cfkkt3g00000gn/C/com.jetbrains.intellij",
		"/private/var/folders/8q/5llq9zx167d0__x3cfkkt3g00000gn/C/com.jetbrains.intellij",

	}
	deleteFiles(append(filePathsForCurrentUser(getCurrentUserName(), filesToRemove), filesToRemove2...))

	installProgram(filename)
}

func installProgram(filename string) {
	fmt.Println("Installing...")
	fmt.Println("Mounting", filename)
	mountImage(filename)
	fmt.Println("Moving to Applications")
	copyToApplications("/Volumes/Intellij\\ IDEA/IntelliJ\\ IDEA.app")
	fmt.Println("Unmounting", filename)
	unmountImage(filename)
	fmt.Printf("%s Installed", filename)
}

func copyToApplications(source string) {
	copyToFolder(source, "/Applications")
}

func copyToFolder(source string, target string) {
	copyCommand := fmt.Sprintf("cp -R %s %s", source, target)
	err := exec.Command("sh", "-c", copyCommand).Run()
	if err != nil {
		return
	}
}

func unmountImage(filename string) {
	err := exec.Command("hdiutil", "detach", filename).Run()
	if err != nil {
		return
	}
}
func ejectImage(filename string) {
	err := exec.Command("hdiutil", "detach", filename).Run()
	if err != nil {
		return
	}
}

func mountImage(filename string) {
	err := exec.Command("hdiutil", "attach", filename).Run()
	if err != nil {
		return
	}
}

func deleteFiles(filePaths []string) {
	for _, path := range filePaths {
		fmt.Println("Removing... " + path)
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Println("Error ", err)
			return
		}
	}

	fmt.Printf("Removed all %d files\n", len(filePaths))
}

func filePathsForCurrentUser(username string, files []string) (paths []string) {
	for _, file := range files {
		path := fmt.Sprintf(file, username)
		paths = append(paths, path)
	}

	return
}

func getCurrentUserName() string {
	current, _ := user.Current()
	return current.Username

}

func downloadFile(err error, filename string, downloadUrl string) {
	err = DownloadFile(filename, downloadUrl)
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded: ", filename)
}

func filterProduct(products []Product, test func(product Product) bool) (res []Product) {
	for _, product := range products {
		if test(product) {
			res = append(res, product)
		}
	}
	return
}

func filterChannel(channels []Channel, test func(channel Channel) bool) (res []Channel) {
	for _, channel := range channels {
		if test(channel) {
			res = append(res, channel)
		}
	}
	return
}

func parseXML(data []byte, out *Products) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	err := decoder.Decode(&out)

	if err != nil {
		return err
	}
	return nil
}

func getXML(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	return body, nil
}

type WriteCounter struct {
	Total uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Total += uint64(n)
	wc.PrintProgress()
	return n, nil
}
func (wc WriteCounter) PrintProgress() {
	// Clear the line by using a character return to go back to the start and remove
	// the remaining characters by filling it with spaces
	fmt.Printf("\r%s", strings.Repeat(" ", 35))

	// Return again and print current status of download
	// We use the humanize package to print the bytes in a meaningful way (e.g. 10 MB)
	fmt.Printf("\rDownloading... %s complete", humanize.Bytes(wc.Total))
}
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	counter := &WriteCounter{}
	_, err = io.Copy(out, io.TeeReader(resp.Body, counter))
	fmt.Print("\n")
	return err
}
