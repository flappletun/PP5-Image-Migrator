package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var wg sync.WaitGroup

var dataPath string
var outPath string

// testing
var imagePath = "resources/Images"

// use
// var imagePath = "/Volumes/pp5/Images"

func main() {

	log.Println("Prompting user for input and output")
	dataPath, outPath = selectFiles()
	log.Println("Input file: " + dataPath)
	log.Println("Output directory: " + outPath)

	// load image filepaths
	log.Println("Collecting image paths from server")
	imagePaths := collectImagePaths(imagePath)
	log.Println("Found", len(imagePaths), "images")

	// load record data from csv file
	log.Println("Loading record data from csv file")
	recordMap := loadData(dataPath)
	log.Println("Loaded", len(recordMap), "records")

	// delete output directory if it exists
	log.Println("Creating output directory")
	info, err := os.Stat(outPath)
	if os.IsNotExist(err) {
		os.MkdirAll(outPath, os.ModePerm)
	} else if info.IsDir() {
		os.RemoveAll(outPath)
		os.MkdirAll(outPath, os.ModePerm)
	}

	// I wonder what the time difference will be
	imageCopyParallel(imagePaths, recordMap)
	//imageCopySequential(imagePaths, recordMap)

}

func imageCopyParallel(imagePaths []string, recordMap map[string]Record) {
	// launch copier goroutines
	var pathChannel = make(chan string, len(imagePaths))
	numGR := 0
	for i := range imagePaths {
		path := filepath.Base(imagePaths[i])
		path = strings.ReplaceAll(path, ".jpg", ".JPG")
		if recordMap[path].ObjectID != "" {
			numGR++
			wg.Add(1)
			go imageCopyWorker(pathChannel, recordMap)
			pathChannel <- imagePaths[i]
		}
	}
	log.Printf("Launched %d goroutines\n", numGR)
	wg.Wait()
	log.Printf("%d goroutines finished\n", numGR)
}

func imageCopySequential(imagePaths []string, recordMap map[string]Record) {
	pathChannel := make(chan string, len(imagePaths))
	wg.Add(len(imagePaths))
	for _, path := range imagePaths {
		pathChannel <- path
		imageCopyWorker(pathChannel, recordMap)
	}
}

func loadData(dataPath string) map[string]Record {
	// open csv file
	dataFile, err := os.Open(dataPath)
	if err != nil {
		panic(err)
	}
	defer dataFile.Close()

	// read csv file
	reader := csv.NewReader(dataFile)
	data, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	// build record structs
	recordMap := make(map[string]Record, 2*len(data))

	/**********************/
	/*  CHANGE THIS LATER */
	objectIDIndex := -1
	imageFileIndex := -1
	/**********************/

	for i, row := range data {
		// find which columns are object ID and image filepath
		if i == 0 {
			for j, column := range row {
				switch column {
				case "OBJECTID":
					objectIDIndex = j
				case "IMAGEFILE":
					imageFileIndex = j
				}
			}
			if objectIDIndex == -1 || imageFileIndex == -1 {
				panic("CSV file is missing required columns")
			}
		} else {
			// add record for each row
			formattedPath := strings.ReplaceAll(row[imageFileIndex], "\\", "/")
			base := filepath.Base(formattedPath)
			recordMap[base] = Record{
				ObjectID: row[objectIDIndex],
				filePath: formattedPath,
			}
		}
	}

	return recordMap
}

func imageCopyWorker(pathChannel chan string, recordMap map[string]Record) {
	var path string
	path = <-pathChannel

	// open the source file
	in, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	// create a destination file
	key := strings.ReplaceAll(filepath.Base(path), ".jpg", ".JPG")
	newName := recordMap[key].ObjectID + ".jpg"
	out, err := os.Create(filepath.Join(outPath, newName))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// copy the contents
	_, err = io.Copy(out, in)
	if err != nil {
		panic(err)
	}

	wg.Done()
}

func collectImagePaths(serverPath string) []string {
	paths := make([]string, 0, 1000)
	err := filepath.Walk(serverPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		} else if strings.HasSuffix(path, ".jpg") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return paths
}
