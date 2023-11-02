package main

import (
	"fmt"
	"math/rand"
	"time"

	"proto/pkg/jsonmsg"
	"proto/pkg/protobufmsg"
)

func compareLength() {
	jsonData, _ := jsonmsg.JsonMsgMarshal("John Doe", 30, "johndoe@example.com")
	pbData, _ := protobufmsg.ProtoBufMsgMarshal("John Doe", 30, "johndoe@example.com")
	fmt.Printf("json len: %d\nprotobuf len: %d\n", len(jsonData), len(pbData))
	fmt.Printf("json data:\n%s\n protobuf data:\n%s\n", jsonData, pbData)
}

func compareTime() {
	// create numItems persons
	numItems := 100000
	persons := make([]jsonmsg.Person, numItems)
	for i := 0; i < numItems; i++ {
		persons[i] = generateRandomPerson()
	}

	// test json
	startTime := time.Now()
	for i := 0; i < numItems; i++ {
		jsonData, _ := jsonmsg.JsonMsgMarshal(persons[i].Name, persons[i].Age, persons[i].Email)
		jsonmsg.JsonMsgUnmarshal(jsonData)
	}
	jsonTime := time.Since(startTime)
	fmt.Printf("json %d times takes %s\n", numItems, jsonTime)

	// test protoBuf
	startTime = time.Now()
	for i := 0; i < numItems; i++ {
		jsonData, _ := protobufmsg.ProtoBufMsgMarshal(persons[i].Name, persons[i].Age, persons[i].Email)
		protobufmsg.ProtoBufMsgUnmarshal(jsonData)
	}
	protoBufTime := time.Since(startTime)
	fmt.Printf("protobuf %d times takes %s\n", numItems, protoBufTime)
}

func generateRandomPerson() jsonmsg.Person {
	return jsonmsg.Person{
		Name:  randomString(10),
		Age:   int32(rand.Intn(100)),
		Email: randomString(15) + "@example.com",
	}
}

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func main() {
	compareLength()
	compareTime()
}
