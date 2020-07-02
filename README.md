# RESTful API 错误码设计

## 1. 前言

现在的软件架构，基本上很多都是采用对外暴露 RESTful API 接口，内部系统通信采用 RPC 协议。因为 RSETful API 接口天生的一些优势：规范、调试友好、易懂，所以通常作为直接面向用户的通信规范。

因为直接面向用户，所以需要有一种合理的方法来满足用户对错误请求的定位和处理。本篇文章就来详细讨论下该如何设计错误码，以及如何提供相应的 go 包来支持。

## 2. 期望的功能

RESTful API 是基于 http 协议的一系列 API 开发规范，http 请求结束后，以下 2 种情况都需要让客户端感知到，以使客户端决定下一步该如何处理：
1. API 请求成功
2. API 请求失败

为了使用户有一个最好的体验，需要有一个比较好的错误码实现方式。这里罗列下，我们在设计错误码时，期望能够实现的功能：

1. 业务 Code 码标识

因为 http code 码有限，并且都是跟 http transport 层相关的 code 码，所以我们希望能有自己的错误 code 码，一方面可以根据需要自行扩展，另一方面也能够精准的定位到具体是哪个错误。同时因为 code 码通常是计算机友好的 10 进制整数，基于 code 码，计算机也可以很方便的进行一些分支处理。

当然了，业务码也要有一定规则，可以通过业务码迅速定位出是哪类错误。

2. 考虑到安全，希望能够对外对内分别展示不同的错误信息

当开发一个对外的系统，业务出错时，需要能够有些机制告诉用户，出了什么错误，如果能够提供一些帮助文档会更好。我们不可能把所有的错误都暴露给外部用户，没必要，也是不安全的，所以也需要有机制能让我们获取到更详细的内部错误信息，这些内部错误信息可能包含一些敏感的数据，不宜对外展示，但可以协助我们进行问题定位。

## 3. 当前错误码设计方式

在业务中，大概有如下三种错误码实现方式（举例：一次请求因为用户账号没有找到而失败）：

### 3.1 不论请求成功失败，始终返回 200 http status code，在 http body 中包含用户账号没有找到的错误信息

例如 Facebook API 的错误 Code 设计，始终返回 200 http status code：

```
{
  "error": {
    "message": "Syntax error \"Field picture specified more than once. This is only possible before version 2.1\" at character 23: id,name,picture,picture",
    "type": "OAuthException",
    "code": 2500,
    "fbtrace_id": "xxxxxxxxxxx"
  }
}
```

采用固定返回 200 http status code 的方式，有其合理性，比如 http code 通常代表 http transport 层的状态信息，当我们收到 http 请求，并返回时，http transport 层是成功的，所以从这个层面上来看，http status 固定为 200，也是合理的。

缺点也很明显：对于每一次请求，我们都要去解析 http body，从中解析出错误码和错误信息，实际上，大部分情况下，我们对于成功的请求，要么直接转发，要买直接解析到某个结构体中。这种方式对性能会有一定的影响，对客户端不友好，不建议这种方式。

### 3.2 返回 http 404 Not Found 错误码，并在 body 中返回简单的错误信息

例如：Twitter API 的错误设计，会根据错误类型，返回合适的 http code，并在 body 中返回错误信息和自定义业务 code。

```
HTTP/1.1 400 Bad Request
x-connection-hash: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
set-cookie: guest_id=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
Date: Thu, 01 Jun 2017 03:04:23 GMT
Content-Length: 62
x-response-time: 5
strict-transport-security: max-age=631138519
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Server: tsa_b

{"errors":[{"code":215,"message":"Bad Authentication data."}]}
```

这种方法是较好的一种方式，通过 http status code 可以使客户端，很方便的知道请求出错了，并且提供客户端一些错误信息供参考。但是仅仅靠这些信息，还不能准确的定位和解决问题。

### 3.3 返回 http 404 Not Found 错误码，并在 body 中返回详细的错误信息

例如微软 Bing API 的错误设计，会根据错误类型，返回合适的 http code，并在 body 中返回详尽的错误信息。

```
HTTP/1.1 400
Date: Thu, 01 Jun 2017 03:40:55 GMT
Content-Length: 276
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Server: Microsoft-IIS/10.0
X-Content-Type-Options: nosniff

{"SearchResponse":{"Version":"2.2","Query":{"SearchTerms":"api error codes"},"Errors":[{"Code":1001,"Message":"Required parameter is missing.","Parameter":"SearchRequest.AppId","HelpUrl":"http\u003a\u002f\u002fmsdn.microsoft.com\u002fen-us\u002flibrary\u002fdd251042.aspx"}]}}
```

这种方式，是比较推荐的一种方式，既能通过 http status code 使客户端方便的知道请求出错，又可以使用户根据返回的信息知道哪里出错，以及如何解决问题。同时，返回了机器友好的业务 code 码，可以使程序在有需要时，进一步进行判断处理。

## 4. 错误码设计建议

综上，可以总结出一套优秀的错误码设计思路：
1. 有区别于 http status code 的业务码，业务码有一定规则，可以通过业务码判断出是哪类错误
2. 请求出错时，可以通过 http status code 感知到请求出错
3. 需要在请求出错时，返回详细的信息，通常包括如下 3 类信息：
    1. 业务 code 码
    2. 错误信息
    3. 参考文档（可选）
4. 返回的错误信息，需要是可以直接展示给用户的安全信息，同时也要有内部更详细的错误信息，方便debug。
5. 返回的数据格式应该是固定的
6. 错误信息要保持简洁，并且提供有用的信息

这里其实还有一些细节需要处理：

1. 业务 code 码设计规范
2. 请求出错时，如何设置 http status code

## 5. 业务 code 码设计

### 5.1 为什么要引入业务 code 码

在实际开发中引入错误码有如下好处：

+ 可以非常方便的定位问题和定位代码行（看到错误码知道什么意思、grep 错误码可以定位到错误码所在行、某个错误类型的唯一标识）
+ 错误码包含一定的信息，通过错误码可以判断出错误级别、错误模块和具体错误信息
+ 业务开发过程中，可能需要判断错误是哪种类型以便做相应的逻辑处理，通过定制的错误可以很容易的做到这点，例如：

```go
    if err == code.ErrBind {
        ...
    }
```

+ Go 中的 http 服务器开发都是引用` net/http` 包，该包中只有 60 个错误码，基本都是跟 http 请求相关的错误码，在一个大型系统中，这些错误码完全不够用，而且这些错误码跟业务没有任何关联，满足不了业务的需求

### 5.2 业务 code 码设计规范

通过研究腾讯云、阿里云、新浪的开放 API，发现新浪的 API code 码设计更合理些，参考新浪的 code 码设计，这里总结出这篇文章推荐的 code 码设计规范：纯数字表示，不同部位代表不同的服务，不同的模块。

**错误代码说明：100101**

+ 10: 服务
+ 01: 某个服务下的某个模块
+ 01: 模块下的错误码序号，每个模块可以注册100个错误

通过 `100101` 可以知道这个错误是 `服务 A` `数据库` 模块下的 `记录没有找到错误`

## 6. 如何设置 http status code

Go `net/http` 包提供了 60 个错误码，大概分为如下 5 类：

+ 1XX - （指示信息）表示请求已接收，继续处理
+ 2XX - （请求成功）表示成功处理了请求的状态代码。
+ 3XX - （请求被重定向）表示要完成请求，需要进一步操作。 通常，这些状态代码用来重定向。
+ 4XX - （请求错误）这些状态代码表示请求可能出错，妨碍了服务器的处理，通常是客户端出错，需要客户端做进一步的处理。
+ 5XX - （服务器错误）这些状态代码表示服务器在尝试处理请求时发生内部错误。 这些错误可能是服务器本身的错误，而不是客户端的问题。

可以看到 http code 有很多种，如果每个 code 都做错误映射，会面临很多问题：

1. 研发同学不太好判断错误属于哪种 http status code，到最后很可能会导致错误或者 http status code 不匹配，变成一种形式
2. 客户端也会疲于应对这么多的 http 错误码

所以，这里建议 http status code 不要太多，基本上只需要如下  3个 http code:

1. 200 - 表示请求成功执行
2. 400 - 表示客户端出问题
3. 500 - 表示服务端出问题

如果觉得这 3 个错误码不够用，最多可以加如下 3 个错误码：
1. 401 - 表示认证失败
2. 403 - 表示授权失败
3. 404 - 表示资源找不到，这里的资源可以是URL或者RESTful资源

将错误码控制在适当的数目，客户端比较好处理和判断，开发也比较容易进行错误码映射。

## 7. 实现

基于以上的讨论，这里给出一种错误码的实现方法。

[marmotedu/errors](https://github.com/marmotedu/errors) - 基于 `github.com/pkg/errors` 增加对 error code 的支持
[marmotedu/sample-code](https://github.com/marmotedu/sample-code) - `github.com/marmotedu/errors` 错误包的错误码实现

### 7.1 错误包 `github.com/marmotedu/errors` 实现

通过在文件 `github.com/marmotedu/errors/errors.go` 中增加新的 `withCode` 结构体，来引入一种新的错误类型，该错误类型，可以记录错误码、stack、cause和具体的错误信息。

```go
type withCode struct {
    err   error
    code  int
    cause error
    *stack
}
```

具体用法如下：

1. 通过 `func WithCode(code int, format string, args ...interface{}) error` 函数来创建新的 `withCode` 类型的错误
2. 通过 `func WrapC(err error, code int, format string, args ...interface{}) error` 来将一个 error 封装成一个 withCode 类型的错误
3. 通过 `func IsCode(err error, code int) bool` 来判断一个 error 链中是包含指定的 code

withCode 最重要的一个方法是 `func (w *withCode) Format(state fmt.State, verb rune)`，在该方法中指定了，不同的打印格式：

+ `%s` 返回可以直接展示给用户的错误信息
+ `%v` alias for `%s`
+ `%-v` 打印出调用栈、错误码、展示给用户的错误信息、展示给研发的错误信息（只展示错误链中，最后一个错误）
+ `%+v` 打印出调用栈、错误码、展示给用户的错误信息、展示给研发的错误信息（展示错误链中的所有错误）
+ `%#-v` JSON 格式打印出调用栈、错误码、展示给用户的错误信息、展示给研发的错误信息（只展示错误链中，最后一个错误）
+ `%#+v` JSON 格式打印出调用栈、错误码、展示给用户的错误信息、展示给研发的错误信息（展示错误链中的所有错误）

> 使用JSON格式打印的日志，可以非常方便的供日志系统解析。

具体使用方式如下：

```go
package main

import (
	"fmt"

	"github.com/marmotedu/errors"
	code "github.com/marmotedu/sample-code"
)

func main() {
	if err := bindUser(); err != nil {
		// %s: Returns the user-safe error string mapped to the error code or the error message if none is specified.
		fmt.Println("====================> %s <====================")
		fmt.Printf("%s\n\n", err)

		// %v: Alias for %s.
		fmt.Println("====================> %v <====================")
		fmt.Printf("%v\n\n", err)

		// %-v: Output caller details, useful for troubleshooting.
		fmt.Println("====================> %-v <====================")
		fmt.Printf("%-v\n\n", err)

		// %+v: Output full error stack details, useful for debugging.
		fmt.Println("====================> %+v <====================")
		fmt.Printf("%+v\n\n", err)

		// %#-v: Output caller details, useful for troubleshooting with JSON formatted output.
		fmt.Println("====================> %#-v <====================")
		fmt.Printf("%#-v\n\n", err)

		// %#+v: Output full error stack details, useful for debugging with JSON formatted output.
		fmt.Println("====================> %#+v <====================")
		fmt.Printf("%#+v\n\n", err)

		// do some business process based on the error type
		if errors.IsCode(err, code.ErrEncodingFailed) {
			fmt.Println("this is a ErrEncodingFailed error")
		}

		if errors.IsCode(err, code.ErrDatabase) {
			fmt.Println("this is a ErrDatabase error")
		}

		// we can also find the cause error
		fmt.Println(errors.Cause(err))
	}
}

func bindUser() error {
	if err := getUser(); err != nil {
		// Step3: Wrap the error with a new error message and a new error code if needed.
		return errors.WrapC(err, code.ErrEncodingFailed, "encoding user 'Lingfei Kong' failed.")
	}

	return nil
}

func getUser() error {
	if err := queryDatabase(); err != nil {
		// Step2: Wrap the error with a new error message.
		return errors.Wrap(err, "get user failed.")
	}

	return nil
}

func queryDatabase() error {
	// Step1. Create error with specified error code.
	return errors.WithCode(code.ErrDatabase, "user 'Lingfei Kong' not found.")
}
```

输出如下：

```
====================> %s <====================
Encoding failed due to an error with the data

====================> %v <====================
Encoding failed due to an error with the data

====================> %-v <====================
encoding user 'Lingfei Kong' failed. - #2 [/home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:53 (main.bindUser)] (100301) Encoding failed due to an error with the data

====================> %+v <====================
encoding user 'Lingfei Kong' failed. - #2 [/home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:53 (main.bindUser)] (100301) Encoding failed due to an error with the data; get user failed. - #1 [/home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:62 (main.getUser)] (100101) Database error; user 'Lingfei Kong' not found. - #0 [/home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:70 (main.queryDatabase)] (100101) Database error

====================> %#-v <====================
[{"caller":"#2 /home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:53 (main.bindUser)","code":100301,"error":"encoding user 'Lingfei Kong' failed.","message":"Encoding failed due to an error with the data"}]

====================> %#+v <====================
[{"caller":"#2 /home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:53 (main.bindUser)","code":100301,"error":"encoding user 'Lingfei Kong' failed.","message":"Encoding failed due to an error with the data"},{"caller":"#1 /home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:62 (main.getUser)","code":100101,"error":"get user failed.","message":"Database error"},{"caller":"#0 /home/lk/workspace/golang/src/github.com/marmotedu/sample-code/examples/main.go:70 (main.queryDatabase)","code":100101,"error":"user 'Lingfei Kong' not found.","message":"Database error"}]

this is a ErrEncodingFailed error
this is a ErrDatabase error
Database error
```

#### 推荐的用法

推荐的用法时，在错误最开始处使用errors.WithCode()创建一个 withCode类型的错误，上层在处理底层错误时，可以用Wrap函数基于该错误，封装新的错误信息，如下：

```go
package main

func main() {
    if err := getUser(); err != nil {
        fmt.Printf("%#+v\n", err)
    }
}

func getUser() error {
	if err := queryDatabase(); err != nil {
		// Step2: Wrap the error with a new error message.
		return errors.Wrap(err, "get user failed.")
	}

	return nil
}

func queryDatabase() error {
	// Step1. Create error with specified error code.
	return errors.WithCode(code.ErrDatabase, "user 'Lingfei Kong' not found.")
}
```

如果要包装的 error 不是用 `github.com/marmotedu/errors` 包创建的，建议用 `errors.WithCode()` 新建一个error。

### 7.2  github.com/marmotedu/sample-code` 错误码实现

`github.com/marmotedu/sample-code` code 码是专门适配于 `github.com/marmotedu/errors` 包的错误码。核心代码位于 `github.com/marmotedu/sample-code/code.go` 文件，该文件创建了，一个 ErrCode 结构体，并实现了 `github.com/marmotedu/errors.Coder` 接口。Coder 接口定义了如下方法：

```go
// Coder defines an interface for an error code detail information.
type Coder interface {
    // HTTP status that should be used for the associated error code.
    HTTPStatus() int

    // External (user) facing error text.
    String() string

    // Reference returns the detail documents for user.
    Reference() string

    // Code returns the code of the coder
    Code() int
}
```

ErrCode 结构体为：

```go
// ErrCode implements `github.com/marmotedu/errors`.Coder interface.
type ErrCode struct {
    // C refers to the code of the ErrCode.
    C int

    // HTTP status that should be used for the associated error code.
    HTTP int

    // External (user) facing error text.
    Ext string

    // Ref specify the reference document.
    Ref string
}
```

我们可以定义一个 ErrCode 类型的code 码，并注册到 errors 包中，错误码包含了如下信息：

1. Int 类型的业务码
2. 对应的 http status code
3. 暴露给外部用户的消息
4. 错误的参考文档

`github.com/marmotedu/sample-code/base.go` 和 `github.com/marmotedu/sample-code/iam.go` 2 个文件中，我们定义了一些错误码。并将这些错误码注册到 errors 包中，在注册的时候，我们会检查 http status code，只允许定义：200、400、401、403、404、500 这 6 个 http 错误码。

### 使用方法

这里举例一个在 `gin` web 框架中使用该错误码的示例：

```go
// Response defines project response format which in marmotedu organization.
type Response struct {
    Code      errors.Code `json:"code,omitempty"`
    Message   string      `json:"message,omitempty"`
    Reference string      `json:"reference,omitempty"`
    Data      interface{} `json:"data,omitempty"`
}

// WriteResponse used to write an error and JSON data into response.
func WriteResponse(c *gin.Context, err error, data interface{}) {
    if err != nil {
        coder := errors.ParseCoder(err)

        c.JSON(coder.HTTPStatus(), Response{
            Code:      coder.Code(),
            Message:   coder.String(),
            Reference: coder.Reference(),
            Data:      data,
        })
    }

    c.JSON(http.StatusOK, Response{Data: data})
}

func GetUser(c *gin.Context) {
    log.Info("get user function called.", "X-Request-Id", requestid.Get(c))
    // Get the user by the `username` from the database.
    user, err := store.Client().Users().Get(c.Param("username"), metav1.GetOptions{})
    if err != nil {
        core.WriteResponse(c, code.ErrUserNotFound.Error(), nil)
        return
    }

    core.WriteResponse(c, nil, user)
}
```

通过 `WriteResponse` 统一处理错误。如果 err != nil 从 error 中解析出 Coder，并调用 Coder 提供的方法，获取错误相关的：http status code、int 类型的业务码、暴露给用户的信息、错误的参考文档链接。如果 err == nil 则返回 200 和 数据。


## 8. 性能

这里我们测试下 `github.com/marmotedu/errors` 的性能，跟 go 标准的 errors 包和 `github.com/pkg/errors` 包的性能进行对比：

```
$  go test -test.bench=BenchmarkErrors -benchtime="3s"
goos: linux
goarch: amd64
pkg: github.com/marmotedu/errors
BenchmarkErrors/errors-stack-10-8         	57658672	        61.8 ns/op	      16 B/op	       1 allocs/op
BenchmarkErrors/pkg/errors-stack-10-8     	 2265558	      1547 ns/op	     320 B/op	       3 allocs/op
BenchmarkErrors/marmot/errors-stack-10-8  	 1903532	      1772 ns/op	     360 B/op	       5 allocs/op
BenchmarkErrors/errors-stack-100-8        	 4883659	       734 ns/op	      16 B/op	       1 allocs/op
BenchmarkErrors/pkg/errors-stack-100-8    	 1202797	      2881 ns/op	     320 B/op	       3 allocs/op
BenchmarkErrors/marmot/errors-stack-100-8 	 1000000	      3116 ns/op	     360 B/op	       5 allocs/op
BenchmarkErrors/errors-stack-1000-8       	  505636	      7159 ns/op	      16 B/op	       1 allocs/op
BenchmarkErrors/pkg/errors-stack-1000-8   	  327681	     10646 ns/op	     320 B/op	       3 allocs/op
BenchmarkErrors/marmot/errors-stack-1000-8         	  304160	     11896 ns/op	     360 B/op	       5 allocs/op
PASS
ok  	github.com/marmotedu/errors	39.200s
```

可以看到 `github.com/marmotedu/errors` 和 `github.com/pkg/errors` 包的性能基本持平：

|package|depth|ns/op|
|----|----|----|
|github.com/pkg/errors|10|1547|
|github.com/marmotedu/errors|10|1772|
|github.com/pkg/errors|100|2881|
|github.com/marmotedu/errors|100|3116|
|github.com/pkg/errors|1000|10646|
|github.com/marmotedu/errors|1000|11896|

## 总结

通过 [marmotedu/errors](https://github.com/marmotedu/errors) 和 [marmotedu/sample-code](https://github.com/marmotedu/sample-code) 实现了我们期望的目标。错误码的设计方式每个业务都有自己的需求，没有一种统一的规范，这里只是一种业务码的设计思路，仅供参考。
