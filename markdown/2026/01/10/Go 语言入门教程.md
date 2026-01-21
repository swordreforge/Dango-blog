# Go 语言入门教程

Go 是一种开源编程语言，它能让构造简单、可靠且高效的软件变得容易。

## 特性

- **简洁**：Go 的语法简洁明了，易于学习
- **高效**：编译速度快，执行效率高
- **并发**：原生支持并发编程
- **安全**：类型安全，内存安全

## Hello World

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

## 变量声明

```go
// 短变量声明
name := "Go"
age := 15

// var 声明
var version string = "1.25"
var isOpen bool = true
```

## 函数

```go
func add(a, b int) int {
    return a + b
}

func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}
```

## 结构体

```go
type Person struct {
    Name string
    Age  int
}

func (p Person) String() string {
    return fmt.Sprintf("%s (%d years)", p.Name, p.Age)
}
```

## 接口

```go
type Speaker interface {
    Speak() string
}

type Dog struct{}

func (d Dog) Speak() string {
    return "Woof!"
}
```

## 并发

```go
func main() {
    go func() {
        fmt.Println("Goroutine 1")
    }()
    
    go func() {
        fmt.Println("Goroutine 2")
    }()
    
    time.Sleep(time.Second)
}
```

## 学习资源

- [Go 官方网站](https://golang.org)
- [Go by Example](https://gobyexample.com)
- [Effective Go](https://golang.org/doc/effective_go)