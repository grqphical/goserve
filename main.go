package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/akamensky/argparse"
	"github.com/fatih/color"
)

// Converts bytes to a nicely formatted string with units
func ConvertSize(bytes int64) string {
	suffixes := []string{"B", "KB", "MB", "GB", "TB"}

	size := float64(bytes)
	i := 0

	for size >= 1024 && i < len(suffixes)-1 {
		size /= 1024
		i++
	}

	return fmt.Sprintf("%.2f%s", size, suffixes[i])
}

// Creates a 404 not found error for the client
func Handle404(conn net.Conn, url string) {
	data, _ := os.ReadFile("templates/404.html")
	status := 404
	response := fmt.Sprintf("HTTP/1.1 %d NOT_FOUND\nContent-Type: text/html; charset=UTF-8\n\n%s", status, data)

	_, err := conn.Write([]byte(response))
	HandleError(err)

	conn.Close()
	red := color.New(color.FgRed).SprintFunc()
	fmt.Printf("%s %s from %s\n", red("404"), url, conn.RemoteAddr())
	return
}

// Handles any errors that may arise during execution
func HandleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// If the user goes to a path that is a diretcory, load a custom HTML page with routes to each
// file in that directory
func MakeDirectoryPage(directoryPath string) string {
	// Load our template HTML file
	template, err := os.ReadFile("templates/directory.html")
	HandleError(err)

	// If the user goes to the server's root, make sure we load files at / and not the name of the current working directory
	dirName := filepath.Base(directoryPath) + "/"
	cwd, err := os.Getwd()
	HandleError(err)
	if directoryPath == cwd {
		dirName = "/"
	}

	// Load the directory name into the page template
	templateString := strings.ReplaceAll(string(template), "$DIRECTORY$", directoryPath)

	files, err := os.ReadDir(directoryPath)
	HandleError(err)

	fileLinks := ""
	// Loop through every file in the directory and add an HTML link to each file to a string
	for _, file := range files {
		fi, err := os.Stat(filepath.Join(directoryPath, file.Name()))
		HandleError(err)

		if file.IsDir() {
			fileLinks += fmt.Sprintf("<p>📁 - <a href='%s%s'>%s</a></p>", dirName, file.Name(), file.Name())
		} else {
			fileLinks += fmt.Sprintf("<p>📄 - <a href='%s%s'>%s</a> - %s</p>", dirName, file.Name(), file.Name(), ConvertSize(fi.Size()))
		}
	}
	// Add the links to the template and return it
	templateString = strings.Replace(templateString, "$LINKS$", fileLinks, 1)
	return templateString
}

func main() {
	// Create cli args
	parser := argparse.NewParser("goserve", "Serves files over http")

	host := parser.String("a", "address", &argparse.Options{Required: false, Help: "host address to run server on", Default: "localhost"})
	port := parser.String("p", "port", &argparse.Options{Required: false, Help: "port to run server on", Default: "8000"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
	}

	fmt.Println("Starting server...")
	// Create our socket server
	server, err := net.Listen("tcp", *host+":"+*port)
	HandleError(err)

	defer server.Close()
	fmt.Printf("Listening on %s at port %s\n", *host, *port)

	for {
		// Handle any connections and pass them to a goroutine
		conn, err := server.Accept()
		HandleError(err)

		go HandleConnection(conn)
	}
}

func HandleConnection(conn net.Conn) {
	// Read the data of the connection
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	HandleError(err)

	// Convert it into an http request object
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buffer)))
	HandleError(err)

	url := request.URL.Path
	status := 200
	response := ""
	cwd, err := os.Getwd()
	HandleError(err)

	filePath := filepath.Join(cwd, url[1:])

	// First check if user request a directory
	// if so return the directory page
	fi, err := os.Stat(filePath)
	if err != nil {
		Handle404(conn, url)
		return
	}

	if fi.IsDir() {
		data := MakeDirectoryPage(filePath)
		response = fmt.Sprintf("HTTP/1.1 %d OK\nContent-Type: text/html\n\n%s", status, data)
	} else {
		// Load the file in the current directory
		data, err := os.ReadFile(filePath)

		if err != nil {
			// If we cant find the file, return a 404 error
			Handle404(conn, url)
			return
		} else {
			// Read the file and generate an HTTP request
			mimeType := mime.TypeByExtension(filepath.Ext(url))
			response = fmt.Sprintf("HTTP/1.1 %d OK\nContent-Type: %s\n\n%s", status, mimeType, data)
		}
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	statusCode := ""

	if status == 404 {
		statusCode = red(fmt.Sprint(status))
	} else if status == 200 {
		statusCode = green(fmt.Sprint(status))
	}

	// Log the connection in the console
	fmt.Printf("%s %s from %s\n", statusCode, url, conn.RemoteAddr())

	// Send the http request and close the connection
	_, err = conn.Write([]byte(response))
	HandleError(err)

	conn.Close()
}
