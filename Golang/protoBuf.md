# ProtoBuf简介/简单使用

### 1. Introduction

ProtoBuf，全称为Protocol Buffers（协议缓冲区），是一种由Google开发的**轻量级数据序列化格式**。它被设计用于高效地序列化结构化数据，使其可在不同平台、不同应用程序之间进行交换和存储。ProtoBuf可以用于多种编程语言，并具有高性能、高可靠性和可扩展性的特点。



### 2. 与JSON比较

在GO中使用`struct`和`json.Marshal`也可以实现数据的序列化和反序列化。但ProtoBuf具有以下几个优势：

1. **更高的性能和更小的数据大小**：ProtoBuf生成的二进制数据比JSON更紧凑，因此占用更小的存储空间，并且在网络传输时需要更少的带宽。同时，ProtoBuf的编解码速度通常比JSON更快，这对于处理大量数据或高性能要求的应用程序非常有利。

2. **跨语言和跨平台支持**：ProtoBuf支持多种编程语言，包括但不限于Go、Java、C++、Python等。这使得你可以在不同的语言和平台之间交换和共享数据，而不必担心兼容性问题。

3. **数据模型的演化**：ProtoBuf具备对数据模型演化的良好支持。当你的消息结构发生变化时，ProtoBuf允许你向前兼容和向后兼容地更新应用程序。旧版本的代码可以与新版本的消息进行互操作，而不会导致严重的错误或数据丢失。这使得应用程序的升级和维护更加灵活和可靠。

4. **显式的消息定义和强类型检查**：ProtoBuf使用明确的消息定义文件来描述数据结构，这使得数据的含义更加明确和可读性更强。此外，使用ProtoBuf进行序列化和反序列化时，会进行强类型检查，减少了因数据格式错误而引发的潜在问题。



### 3. go使用protoBuf通信

首先看下目录结构，后面文件创建按该目录来

```shell
.
├── data
├── go.mod
├── go.sum
├── main.go
└── message
    ├── message.pb.go
    └── message.proto
```

#### 3.1 安装protobuf-compiler

```shell
yum install protobuf-compiler
protoc --version
```

#### 3.2 定义编译protoBuf

定义一个简单的消息类型`Person`，`vim message.proto`写入

```protobuf
syntax = "proto3";

option go_package = "./";

package message;

message Person {
  string name = 1;
  int32 age = 2;
}
```

编译，编译成功后可生成`message.pb.go`一个文件

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
protoc --go_out=. message.proto
```

#### 3.3 go模块

go主程序，`vim main.go`

- send：生成一个Person，序列化为protoBuf，写入data文件
- recv：从data文件读出，反序列化后打印输出

```go
package main

import (
        "fmt"
        "io/ioutil"
        "log"

        "github.com/golang/protobuf/proto"

        "proto/message"
)

func send() {
        person := &message.Person{
                Name: "Alice",
                Age:  30,
        }

        data, err := proto.Marshal(person)
        if err != nil {
                log.Fatal("failed to marshal:", err.Error())
        }

        err = ioutil.WriteFile("data", data, 0644)
        if err != nil {
                log.Fatal("failed to write:", err.Error())
        }

        fmt.Println("success send message")
}

func recv() {
        data, err := ioutil.ReadFile("data")
        if err != nil {
                log.Fatal("failed to read data, error: ", err.Error())
        }

        person := &message.Person{}
        err = proto.Unmarshal(data, person)
        if err != nil {
                log.Fatal("failed to umarshal data:", err.Error())
        }

        fmt.Println("success recv message", person)
}

func main() {
        send()
        recv()
}
```

编译运行：

```shell
[root@iZuf6ffajhlqk1snuknbs5Z protoBuf]# go build main.go
# 成功执行
[root@iZuf6ffajhlqk1snuknbs5Z protoBuf]# ./main
success send message
success recv message name:"Alice" age:30
```

![image-20231102225052428](D:\code\FirstTry\Golang\image-20231102225052428.png)
