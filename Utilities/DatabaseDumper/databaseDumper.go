package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
	"github.com/PaulSnow/factom2d/database/databaseOverlay"
	"github.com/PaulSnow/factom2d/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

func main() {
	fmt.Println("Usage:")
	fmt.Println("BlockExtractor level/bolt DBFileLocation")
	fmt.Println("Database will be dumped into a json-formatted text file")

	if len(os.Args) < 3 {
		fmt.Println("\nNot enough arguments passed")
		os.Exit(1)
	}
	if len(os.Args) > 3 {
		fmt.Println("\nToo many arguments passed")
		os.Exit(1)
	}

	levelBolt := os.Args[1]

	if levelBolt != level && levelBolt != bolt {
		fmt.Println("\nFirst argument should be `level` or `bolt`")
		os.Exit(1)
	}
	path := os.Args[2]

	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == bolt {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}

	err = ExportDatabaseJSON(dbase, true)
	if err != nil {
		panic(err)
	}
}

func ExportDatabaseJSON(db interfaces.IDatabase, convertNames bool) error {
	fmt.Printf("Exporting the database\n")
	if db == nil {
		return nil
	}
	buckets, err := db.ListAllBuckets()
	if err != nil {
		return err
	}
	answer := map[string]interface{}{}
	for _, bucket := range buckets {
		m := map[string]interface{}{}
		data, keys, err := db.GetAll(bucket, new(primitives.ByteSlice))
		if err != nil {
			return err
		}
		for i, key := range keys {
			m[fmt.Sprintf("%x", key)] = data[i]
		}
		if convertNames == true {
			answer[KeyToName(bucket)] = m
		} else {
			answer[fmt.Sprintf("%x", bucket)] = m
		}
	}

	data, err := primitives.EncodeJSON(answer)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	data = out.Next(out.Len())
	/*
		dir := be.DataStorePath
		if dir != "" {
			if FileNotExists(dir) {
				err := os.MkdirAll(dir, 0777)
				if err == nil {
					fmt.Println("Created directory " + dir)
				} else {
					return err
				}
			}
		}
		if dir != "" {
			dir = dir + "/db.txt"
		} else {
			dir = "db.txt"
		}*/
	dir := "db.txt"
	err = ioutil.WriteFile(dir, data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func KeyToName(key []byte) string {
	name, ok := databaseOverlay.ConstantNamesMap[string(key)]
	if ok == true {
		return name
	}
	return fmt.Sprintf("%x", key)
}
