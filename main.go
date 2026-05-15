package main

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var outPath = "resources/out"
var wg sync.WaitGroup

func main() {

	// load image filepaths
	imagePaths := collectImagePaths("resources/Yoakum")
	println("Found", len(imagePaths), "images")

	// load record data from csv file
	recordMap := loadData("resources/test_data.csv")
	println("Loaded", len(recordMap), "records")

	// delete output directory if it exists
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
	for range imagePaths {
		wg.Add(1)
		go imageCopyWorker(pathChannel, recordMap)
	}

	// give the workers the paths to copy
	for _, path := range imagePaths {
		pathChannel <- path
	}
	wg.Wait()
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
	recordMap := make(map[string]Record, len(data))

	/**********************/
	/*  CHANGE THIS LATER */
	objectIDIndex := 0
	imageFileIndex := 1
	/**********************/

	for i, row := range data {
		// find which columns are object ID and image filepath
		if i == 0 {
			for j, column := range row {
				println("Column", j, "is", column)
				switch column {
				case "objectid":
					objectIDIndex = j
				case "imagefile":
					imageFileIndex = j
				}
			}
			if objectIDIndex == -1 || imageFileIndex == -1 {
				panic("CSV file is missing required columns")
			}
		} else {
			// add record for each row
			recordMap[row[imageFileIndex]] = Record{
				ObjectID: row[objectIDIndex],
				filePath: row[imageFileIndex],
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
	newName := recordMap[filepath.Base(path)].ObjectID + ".jpg"
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
