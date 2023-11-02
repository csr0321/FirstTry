package protobufmsg

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

func ProtoBufMsgMarshal(name string, age int32, email string) ([]byte, error) {
	protoBufData, err := proto.Marshal(&Person{
		Name:  name,
		Age:   age,
		Email: email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protoBuf %s", err.Error())
	}
	return protoBufData, nil
}

func ProtoBufMsgUnmarshal(protoBufData []byte) (res *Person, err error) {
	res = &Person{}
	err = proto.Unmarshal(protoBufData, res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal protoBuf %s", err.Error())
	}
	return res, err
}
