package jsonmsg

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name  string `json:"name"`
	Age   int32  `json:"age"`
	Email string `json:"email"`
}

func JsonMsgMarshal(name string, age int32, email string) ([]byte, error) {
	jsonData, err := json.Marshal(&Person{
		Name:  name,
		Age:   age,
		Email: email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json %s", err.Error())
	}
	return jsonData, nil
}

func JsonMsgUnmarshal(jsonData []byte) (res *Person, err error) {
	res = &Person{}
	err = json.Unmarshal(jsonData, res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %s", err.Error())
	}
	return res, err
}
