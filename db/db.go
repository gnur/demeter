package db

import "github.com/asdine/storm"

var (
	//Conn is the shared pointer to the initialized database
	Conn *storm.DB
)
