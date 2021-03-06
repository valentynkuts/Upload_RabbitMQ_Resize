package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/streadway/amqp"
)

const (
	HOST        = "http://localhost"
	PORT        = ":8080"
	URL_UPLOAD  = "/upload"
	IMAGES_PATH = "../temp_images"
	BROKER_URL  = "amqp://guest:guest@localhost:5672/"
)

func myHtmlForm(w http.ResponseWriter, r *http.Request) {

	tmpl, _ := template.ParseFiles("index.html")
	url := URL_UPLOAD
	tmpl.Execute(w, url)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	//parse input, type multipart/form-data
	r.ParseMultipartForm(10 << 20)

	// retrieve file from posted form-data
	file, handler, err := r.FormFile("myFile")

	if err != nil {
		fmt.Println("Error Retrieving file from form-data")
		fmt.Println(err)
		return
	}

	defer file.Close()

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	//to get name and extension
	arr_name_ext := strings.Split(handler.Filename, ".")

	file_id_create := fmt.Sprintf("%s-*.%s", arr_name_ext[0], arr_name_ext[1])

	//Creates a new temporary file
	//filename is generated by taking pattern and adding a random string
	tempFile, err := ioutil.TempFile(IMAGES_PATH, file_id_create)

	if err != nil {
		fmt.Println(err)
		return
	}

	info_path := tempFile.Name()
	fmt.Printf("tempFile.Name(): %+v\n", info_path)
	arr_dir_filename := strings.Split(info_path, "/")
	len := len(arr_dir_filename)
	file_id := arr_dir_filename[len-1]

	fmt.Printf("File ID: %+v\n", file_id)

	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)

	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	const link = `<p><a href="/">BACK to upload file</a></p>`
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, link)

	//return whether or not this has been successful
	fmt.Fprintf(w, "Successfully Uploaded File\n")

	producer(file_id)

}

func setupRoutes() {
	http.HandleFunc("/", myHtmlForm)
	http.HandleFunc(URL_UPLOAD, uploadFile)
	http.ListenAndServe(PORT, nil)

}

func producer(file_id string) {

	//The Dial function connects to a server
	conn, err := amqp.Dial(BROKER_URL)
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		panic(err)
	}

	// Let's start by opening a channel to our RabbitMQ instance
	// over the connection we have already established
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	// with this channel open, we can then start to interact
	// with the instance and declare Queues that we can publish and
	// subscribe to
	q, err := ch.QueueDeclare(
		"ResizeImage",
		false,
		false,
		false,
		false,
		nil,
	)
	// We can print out the status of our Queue here
	// this will information like the amount of messages on
	// the queue
	fmt.Println(q)
	// Handle any errors if we were unable to create the queue
	if err != nil {
		fmt.Println(err)
	}

	// attempt to publish a message to the queue!
	err = ch.Publish(
		"",
		"ResizeImage",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(file_id),
		},
	)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Published Message to Queue")
}

func main() {
	fmt.Println("Go File Upload")
	setupRoutes()

}
