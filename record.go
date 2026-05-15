package main

type Record struct {
	ObjectID string
	filePath string
}

func NewRecord(objectID, filePath string) Record {
	return Record{
		ObjectID: objectID,
		filePath: filePath,
	}
}
