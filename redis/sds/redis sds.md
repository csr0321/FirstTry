## Redis - SDS 数据结构

今天来看下redis sds的[源码](https://github.com/redis/redis/tree/7.2)，目前redis最新稳定版本是`7.2.2`。

阅读代码：`commit 7f4bae817614988c43c3024402d16edcbf3b3277 (HEAD -> 7.2, tag: 7.2.3, origin/7.2)`

### 1. SDS简介

redis有五种基础数据结构string、list、hash、set、zset。其中string是 Redis 中最简单的数据类型。它可以包含任何数据，包括文本、二进制数据等。

string的底层数据结构为SDS（Simple Dynamic String，简单动态字符串）。主要作用有两个：

- 实现字符串对象（StringObject）；
- 在 Redis 程序内部用作 `char*` 类型的替代品；

在 Redis 中， 客户端传入服务器的协议内容、 aof 缓存、 返回给客户端的回复， 等等， 这些重要的内容都是由 sds 类型来保存的。



### 2. SDS 结构

在c语言中，没有原生支持的字符串类型，通常使用`char *`表示，`char *`指向的地址为字符串的第一个字符，其中字符数组的末尾通常包含一个空字符(`\0`)来表示字符串的结束。如果想获取字符串的长度，需要遍历到`\0`，gcc源码中的strlen：

```c
__attribute__ ((__noinline__))
__SIZE_TYPE__
strlen (const char *s)
{
  __SIZE_TYPE__ i;
  i = 0;
  while (s[i] != 0)
    i++;
  return i;
}
```

- 所以char*获取字符串长度的时间复杂度为O(n)，并不能像c++的`string.length()`和go的`len(string)`那样O(1)获取长度；
- 对字符串append会引起内存的`realloc()`，开销更大。
- Redis 除了处理 C 字符串之外， 还需要处理单纯的字节数组， 以及服务器协议等内容， 所以为了方便起见， Redis 的字符串表示还应该是二进制安全的： 程序不应对字符串里面保存的数据做任何假设， 数据可以是以 `\0` 结尾的 C 字符串， 也可以是单纯的字节数组， 或者其他格式的数据。

综上所述，Redis设计了新的字符串SDS替代了`char*`。SDS的结构体在`src\sds.h`中定义，有五个struct的定义：

```c
struct __attribute__ ((__packed__)) sdshdr5 {
    unsigned char flags; /* 3 lsb of type, and 5 msb of string length */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr8 {
    uint8_t len; /* used */
    uint8_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr16 {
    uint16_t len; /* used */
    uint16_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr32 {
    uint32_t len; /* used */
    uint32_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr64 {
    uint64_t len; /* used */
    uint64_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
```

除了`sdshdr5`格式不太一样(sdshdr5后面再讲)，其他四个基本一致，差异点只在`len`和`alloc`的数据类型。再把分支切回`3.0`，看一下以前的sds：

```c++
struct sdshdr {
    unsigned int len;
    unsigned int free;
    char buf[];
};
```

发现`7.2`较`3.0`：

- 字段差异：
  - `7.2`多了一个flags用来标识是五种其一的sds，`3.0`只有一种sds
  - `7.2`使用alloc标识总长度，`3.0`用free标识未使用长度
- 类型差异：`7.2`长度标识类型有四种`uint8`、`uint16`、`uint32`、`uint64`，`3.0`只有`uint32`类型的长度标识
- 编译选项：`7.2`使用`__attribute__ ((packed)) `取消了字节对齐，使空间更紧凑

综上所述，其实就一句话：**为了省内存**。取消了字节对齐，在字符串没那么长时候采用了8位 or 16位长度标识，减少字符头的内存占用。（32位程序如图所示）

![image-20231107220529935](D:\code\FirstTry\redis\sds\image-20231107220529935.png)

### 3. SDS 操作

- 探讨 Redis 源码中常用的 SDS 操作，如创建、销毁、追加、裁剪等。
- 提供示例代码和相应函数的源码片段。

#### 3.1 sdslen() 长度获取

获取sds长度buf长度的，总体分为三步

1. 获取flags：`unsigned char flags = s[-1];` 获取 sds 结构体中的 `flags` 字段，通过将指针 `s` 向前移动一个字节（`s[-1]`），来获取 `flags` 的值（在嵌入式领域这样处理消息的常用手法，通常会将指针指向消息头或者payload，向前偏移来获取上层消息或消息头）
2. 根据flags判断是哪种SDS：`switch(flags&SDS_TYPE_MASK)` 用掩码运算来提取 `flags` 中的类型信息， `flags` 的低 3 位用来表示 SDS 类型，例如`flags=3`，掩码`SDS_TYPE_MASK`为`0b0111`，可以得 `0b0011 & 0b0111 = 0b0011`为SDS_TYPE_32
3. 根据SDS类型获取len字段：在 `switch` 语句中，根据不同的 SDS 类型，代码会访问相应类型的 `sdshdr` 结构体并获取 `len` 字段的值。例如，对于 `SDS_TYPE_8`，使用 `SDS_HDR(8, s)->len` 获取 `len` 字段

```c
#define SDS_TYPE_5  0   // 0b0000
#define SDS_TYPE_8  1   // 0b0001
#define SDS_TYPE_16 2   // 0b0010
#define SDS_TYPE_32 3   // 0b0011
#define SDS_TYPE_64 4   // 0b0100
#define SDS_TYPE_MASK 7 // 0b0111
#define SDS_HDR(T,s) ((struct sdshdr##T *)((s)-(sizeof(struct sdshdr##T)))) // 获取对应结构体
#define SDS_TYPE_5_LEN(f) ((f)>>SDS_TYPE_BITS)

static inline size_t sdslen(const sds s) {
    unsigned char flags = s[-1];  // s指向的是buf，所以向后一个指针是flags
    switch(flags&SDS_TYPE_MASK) {
        case SDS_TYPE_5:
            return SDS_TYPE_5_LEN(flags);
        case SDS_TYPE_8:
            return SDS_HDR(8,s)->len;
        case SDS_TYPE_16:
            return SDS_HDR(16,s)->len;
        case SDS_TYPE_32:
            return SDS_HDR(32,s)->len;
        case SDS_TYPE_64:
            return SDS_HDR(64,s)->len;
    }
    return 0;
}
```

PS：

到处都在说SDS的最大buffer长度为512MB，包括官方文档，但是源码中的sdsTypeMaxSize函数最大限制为`(1ll<<32) - 1`，在64位系统下远远大于512MB。

```C
static inline size_t sdsTypeMaxSize(char type) {
    if (type == SDS_TYPE_5)
        return (1<<5) - 1;
    if (type == SDS_TYPE_8)
        return (1<<8) - 1;
    if (type == SDS_TYPE_16)
        return (1<<16) - 1;
#if (LONG_MAX == LLONG_MAX)
    if (type == SDS_TYPE_32)
        return (1ll<<32) - 1;
#endif
    return -1; /* this is equivalent to the max SDS_TYPE_64 or SDS_TYPE_32 */
}
```

经查阅资料发现，实际在`set`或者`append`时候在`checkStringLength()`处检查长度，长度限制`server.proto_max_bulk_len`由`createLongLongConfig`赋值为512MB

```c
static int checkStringLength(client *c, long long size, long long append) {
    if (mustObeyClient(c))
        return C_OK;
    /* 'uint64_t' cast is there just to prevent undefined behavior on overflow */
    long long total = (uint64_t)size + append;
    /* Test configured max-bulk-len represending a limit of the biggest string object,
     * and also test for overflow. */
    if (total > server.proto_max_bulk_len || total < size || total < append) {
        addReplyError(c,"string exceeds maximum allowed size (proto-max-bulk-len)");
        return C_ERR;
    }
    return C_OK;
}
```

```c
createLongLongConfig("proto-max-bulk-len", NULL, DEBUG_CONFIG | MODIFIABLE_CONFIG, 1024*1024, LONG_MAX, server.proto_max_bulk_len, 512ll*1024*1024, MEMORY_CONFIG, NULL, NULL), /* Bulk request max size */
```



#### 3.2 sdsnewlen 创建sds

sds创建流程：

1. 根据初始化长度 `initlen` 确定 sds 的类型 `type`，`sdsReqType()`根据`initlen` 返回满足sds len的最大长度类型，例如`initlen` 大于等于 `1<<16` 但小于 `1<<32`，返回SDS_TYPE_32
2. 如果 `type` 是 `SDS_TYPE_5` 并且 `initlen` 为 0，则将 `type` 设置为 `SDS_TYPE_8`
3. 申请消息内存，`trymalloc`调用的`s_trymalloc_usable`和`s_malloc_usable`的区别在于try有一个OOM的捕获`if (!ptr) zmalloc_oom_handler(size);`
4. 初始化 sds 数据区域，将数据区域初始化为零
5. 根据 sds 类型，设置 sds 头部和标志字节
6. 最后，将数据区域的最后一个字节设置为`'\0'`终结符，以确保 sds 表现为 C 字符串
7. 返回 sds 结构的指针

```c
sds _sdsnewlen(const void *init, size_t initlen, int trymalloc) {
    void *sh;
    sds s;
    char type = sdsReqType(initlen); // 1. 根据需要初始化的长度设置sds的类型
    if (type == SDS_TYPE_5 && initlen == 0) type = SDS_TYPE_8; // 2. 如果是sdshdr5默认设置为sdshdr8
    int hdrlen = sdsHdrSize(type);
    unsigned char *fp; /* flags pointer. */
    size_t usable;

    assert(initlen + hdrlen + 1 > initlen); /* Catch size_t overflow */
    sh = trymalloc?
        s_trymalloc_usable(hdrlen+initlen+1, &usable) :
        s_malloc_usable(hdrlen+initlen+1, &usable); // 3. 申请消息内存
    if (sh == NULL) return NULL;
    if (init==SDS_NOINIT)
        init = NULL;
    else if (!init)
        memset(sh, 0, hdrlen+initlen+1); // 4. 数据区域初始化为零
    s = (char*)sh+hdrlen;
    fp = ((unsigned char*)s)-1;
    usable = usable-hdrlen-1;
    if (usable > sdsTypeMaxSize(type))
        usable = sdsTypeMaxSize(type);
    switch(type) {
        case SDS_TYPE_5: {
            *fp = type | (initlen << SDS_TYPE_BITS);
            break;
        }
        case SDS_TYPE_8: {
            ...
        }
        case SDS_TYPE_16: {
            ...
        }
        case SDS_TYPE_32: { // 5.设置len、alloc、flags
            SDS_HDR_VAR(32,s);
            sh->len = initlen;
            sh->alloc = usable;
            *fp = type;
            break;
        }
        case SDS_TYPE_64: {
           ...
        }
    }
    if (initlen && init)
        memcpy(s, init, initlen);
    s[initlen] = '\0'; // 6.将数据区域的最后一个字节设置为'\0'终结符
    return s; // 7.返回 sds 结构的指针
}
```



#### 3.3 sdsMakeRoomFor 扩容机制

扩容函数`_sdsMakeRoomFor`

1. 获取当前 sds 字符串的可用空间大小 `avail`，如果可用空间大于所需空间`addlen`则直接返回，
2. 计算新的字符串长度 `newlen`，等于当前长度`len` +所需空间 `addlen`
3. 如果 `newlen` 小于 `SDS_MAX_PREALLOC(1MB=1024*1024)`，且 `greedy` 参数为 1，那么将 `newlen` 增大到原来的两倍，否则增加 `SDS_MAX_PREALLOC(1MB)`，计算机领域很多内存分配都是这样做的，在一个门限之内*2，门限之外线性增长
4. 根据新长度`newlen` 确定sds是否要改变长度类型
5. 检查新的类型 `type` 是否与原类型 `oldtype` 相同，如果相同，尝试使用 `realloc` 来调整内存块的大小，同时保留之前的数据。分配成功后，将 `s` 指向新的数据块，并将 `usable` 设置为可用的内存大小
6. 如果新的类型 `type` 与原类型 `oldtype` 不同，需要分配新的内存块并移动数据。
   - 分配新的内存块 `newsh`，并将数据从原 sds 复制到新的内存块。
   - 释放原来的内存块 `sh`。
   - 将 `s` 指向新的数据块，设置新的类型 `type`，并更新 sds 的长度
7. 设置新的`alloc`，并返回更新后的 sds 指针

```c
sds _sdsMakeRoomFor(sds s, size_t addlen, int greedy) {
    void *sh, *newsh;
    size_t avail = sdsavail(s);
    size_t len, newlen, reqlen;
    char type, oldtype = s[-1] & SDS_TYPE_MASK;
    int hdrlen;
    size_t usable;

    /* Return ASAP if there is enough space left. */
    if (avail >= addlen) return s; // 1. 剩余空间充足

    len = sdslen(s);
    sh = (char*)s-sdsHdrSize(oldtype);
    reqlen = newlen = (len+addlen); // 2. 计算新长度
    assert(newlen > len);   /* Catch size_t overflow */
    if (greedy == 1) {
        if (newlen < SDS_MAX_PREALLOC)
            newlen *= 2;
        else
            newlen += SDS_MAX_PREALLOC;
    }

    type = sdsReqType(newlen); // 4. 获取sds类型

    /* Don't use type 5: the user is appending to the string and type 5 is
     * not able to remember empty space, so sdsMakeRoomFor() must be called
     * at every appending operation. */
    if (type == SDS_TYPE_5) type = SDS_TYPE_8;

    hdrlen = sdsHdrSize(type);
    assert(hdrlen + newlen + 1 > reqlen);  /* Catch size_t overflow */
    if (oldtype==type) {
        newsh = s_realloc_usable(sh, hdrlen+newlen+1, &usable); // 5. 同类型realloc
        if (newsh == NULL) return NULL;
        s = (char*)newsh+hdrlen;
    } else {
        /* Since the header size changes, need to move the string forward,
         * and can't use realloc */
        newsh = s_malloc_usable(hdrlen+newlen+1, &usable); // 6. 不同类型malloc、memcpy，free老内存
        if (newsh == NULL) return NULL;
        memcpy((char*)newsh+hdrlen, s, len+1);
        s_free(sh);
        s = (char*)newsh+hdrlen;
        s[-1] = type;
        sdssetlen(s, len);
    }
    usable = usable-hdrlen-1;
    if (usable > sdsTypeMaxSize(type))
        usable = sdsTypeMaxSize(type);
    sdssetalloc(s, usable); // 7. 设置alloc并返回指针
    return s;
}
```



**参考文献：**

1. [Redis 数据结构：Simple Dynamic Strings (SDS)](https://redisbook.readthedocs.io/en/latest/internal-datastruct/sds.html)
2. [Redis 数据结构与命令详解](https://www.51cto.com/article/700992.html)