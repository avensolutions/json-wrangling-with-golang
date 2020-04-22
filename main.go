package main

// jsontogo.exe select|describe <file> <collection> <cols>
// ./jsontogo.exe select ./zones.json items
// ./jsontogo.exe describe ./zones.json items

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	cmdLineArgs := os.Args[1:]

	if len(cmdLineArgs) < 3 {
		fmt.Print("Program requires three arguments (operation [select|describe], json_file, collection)")
		return
	}

	operation := cmdLineArgs[0]
	json_file := cmdLineArgs[1]
	collection := cmdLineArgs[2]

	var maxfieldlen int = 10

	//if len(cmdLineArgs) > 3 {
	//	maxfieldlen = cmdLineArgs[3]
	//}

	if operation != "select" && operation != "describe" {
		log.Printf("[%s] is not a valid operation - valid operations include select, describe", operation)
		return
	}

	log.Printf("Reading JSON data from %s, collection [%s]", json_file, collection)

	body, err := ioutil.ReadFile(json_file)
	check(err)

	// validate JSON
	if json.Valid(body) {
		log.Printf("Valid JSON")
	} else {
		log.Printf("ERROR: Invalid JSON")
		return
	}

	// declare an empty interface used to decode the json object
	var data interface{}
	var colldata interface{}

	log.Print("Decoding response...")

	// decode json to the empty interface (data)
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Print("ERROR: Invalid JSON Object")
		return
	}

	// get object type using reflection
	log.Printf("Decoded Object Type (using reflection): %s", reflect.TypeOf(data))

	objType := getObjectType(data)

	// check if JSON object is a map and look for the collection key
	if objType == "map" {
		log.Printf("Searching for %s in decoded JSON object", collection)
		rec := data.(map[string]interface{})
		for key, value := range rec {
			if key == collection {
				log.Printf("%s key found", collection)
				colldata = value
				break
			}
		}
		if colldata == nil {
			log.Printf("ERROR: %s key not found", collection)
			return
		}
	} else {
		log.Printf("ERROR: Not a map")
		return
	}

	// confirm collection is an array
	if getObjectType(colldata) != "array" {
		log.Printf("ERROR: %s is not an array", collection)
		return
	}

	if operation == "select" {
		// Recurse object
		RecurseCollection(colldata.([]interface{}), maxfieldlen)
	} else {
		// Describe collection
		DescribeCollection(colldata.([]interface{}))
	}
}

func getObjectType(data interface{}) string {
	// use a type switch to return object type
	objType := "unknown"
	switch v := data.(type) {
	case int:
		fmt.Println("int:", v)
	case float64:
		fmt.Println("float64:", v)
	case []interface{}:
		objType = "array"
	case map[string]interface{}:
		objType = "map"
	case string:
		objType = "string"
	default:
		fmt.Println("unknown")
	}
	return objType
}

func DescribeCollection(colldata []interface{}) {
	var headers []string
	var tabledata [][]string
	var recorddata []string

	headers = append(headers, "name")
	headers = append(headers, "type")

	// get first map in array
	for key, value := range colldata[0].(map[string]interface{}) {
		recorddata = nil
		recorddata = append(recorddata, key)
		recorddata = append(recorddata, getObjectType(value))
		tabledata = append(tabledata, recorddata)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range tabledata {
		table.Append(v)
	}
	table.Render() // Send output
}

func RecurseCollection(colldata []interface{}, maxfieldlen int) {
	var headers []string
	var tabledata [][]string
	var recorddata []string

	headers = append(headers, "id")
	headers = append(headers, "name")

	// get remaining column headers
	for key, _ := range colldata[0].(map[string]interface{}) {
		if key != "id" && key != "name" {
			headers = append(headers, key)
		}
	}

	for _, v := range colldata {
		// get each map in the array
		recorddata = nil
		for i := range headers {
			var maxlen int
			stringvalue := fmt.Sprintf("%v", v.(map[string]interface{})[headers[i]])
			if headers[i] == "id" || headers[i] == "name" {
				// dont truncate the id or name field
				recorddata = append(recorddata, stringvalue)
			} else {
				if maxfieldlen >= len(headers[i]) {
					maxlen = maxfieldlen
				} else {
					maxlen = len(headers[i])
				}

				if len(stringvalue) > maxlen {
					recorddata = append(recorddata, stringvalue[0:maxlen])
				} else {
					recorddata = append(recorddata, stringvalue)
				}
			}
		}
		tabledata = append(tabledata, recorddata)
		break
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range tabledata {
		table.Append(v)
	}
	table.Render() // Send output
}
