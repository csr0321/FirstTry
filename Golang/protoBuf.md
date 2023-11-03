# ProtoBuf简介/简单使用

### 1. 简介

ProtoBuf，全称为 Protocol Buffers（协议缓冲区），是一种由 Google 开发的**轻量级数据序列化格式**。设计用于高效地序列化结构化数据，使其可在不同平台、不同应用程序之间进行交换和存储。ProtoBuf 可以用于多种编程语言，并具有高性能、高可靠性和可扩展性的特点。

ProtoBuf 在业务场景中经常用于传输数据。

### 2. 与 JSON 比较

在 Go 语言中，可以使用 `struct` 和 `json.Marshal` 来实现数据的序列化和反序列化。ProtoBuf 具有以下几个优势：

1. 更高的性能和更小的数据大小：ProtoBuf 生成的二进制数据**比 JSON 更紧凑**，因此占用更小的存储空间，传输时需要更少的带宽。ProtoBuf 的编解码速度通常比 JSON 更快。
2. 跨语言和跨平台支持：支持多种编程语言，包括但不限于 Go、Java、C++、Python 等。可以在不同的语言和平台之间交换和共享数据，不必担心兼容性问题。
3. 显式消息定义和强类型检查：ProtoBuf 使用明确的消息定义文件来描述数据结构，让数据的含义更加明确和可读性更强。使用 ProtoBuf 进行序列化和反序列化时，会进行强类型检查，减少因数据格式错误而引发的潜在问题。



### 3. 示例代码和性能比较

#### 3.1 拉取代码

从以下 GitHub 仓库中获取示例代码：[GitHub 仓库链接 ](https://github.com/fuckusedname/FirstTry/tree/main/Golang/protoBuf)目录结构如下：

```shell
# tree
.
├── data
├── go.mod
├── go.sum
├── main
├── main.go
└── pkg
    ├── jsonmsg
    │   └── jsonmsg.go
    └── protobufmsg
        ├── message.pb.go
        ├── message.proto
        └── protobufmsg.go

3 directories, 9 files
```

#### 3.2 安装protobuf-compiler

```shell
yum install protobuf-compiler
protoc --version
```

#### 3.3 定义消息编译protoBuf

在 `pkg/jsonmsg/jsonmsg.go` 中定义了一个 JSON 类型的 `Person struct`，示例代码如下：

```go
type Person struct {
	Name  string `json:"name"`
	Age   int32  `json:"age"`
	Email string `json:"email"`
}
```

在 `pkg/jsonmsg/jsonmsg.go` 中定义了一个 JSON 类型的 `Person struct`，示例代码如下：

```protobuf
syntax = "proto3";

option go_package = "./";

package protobufmsg;

message Person {
  string name = 1;
  int32 age = 2;
  string email=3;
}
```

编译，编译成功后可生成`message.pb.go`一个文件

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
protoc --go_out=. message.proto
```

#### 3.3 比对序列化后二进制长度

`main.go`文件里面有一个`compareLength()`函数用来比对相同内容序列化后的长度。

```
func compareLength() {
	jsonData, _ := jsonmsg.JsonMsgMarshal("John Doe", 30, "johndoe@example.com")
	pbData, _ := protobufmsg.ProtoBufMsgMarshal("John Doe", 30, "johndoe@example.com")
	fmt.Printf("json len: %d\nprotobuf len: %d\n", len(jsonData), len(pbData))
	fmt.Printf("json data:\n%s\n protobuf data:\n%s\n", jsonData, pbData)
}
```

可以看到，json序列化后长度为58 B，protoBuf为33 B，**protoBuf仅为json的56.9%，内存占用几乎少了一半**。

protoBuf除了各别分隔符，没有key值和各种符号，**json拥有较多冗余数据**。

![image-20231103220925209](D:\code\FirstTry\Golang\image-20231103220925209.png)

#### 3.4 比对序列化和反序列化的耗时

在main.go中，还有一个函数`compareTime()`用作比对序列化和反序列化耗时，生成10w个随机数据，然后分别进行json和protoBuf的序列化和反序列化，protoBuf序列化耗时仅为json的1/3。

![image-20231103221633737](D:\code\FirstTry\Golang\image-20231103221633737.png)
