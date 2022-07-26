package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	reader := strings.NewReader("Go语言中文网")
	p := make([]byte, 6)
	n, err := reader.ReadAt(p, 2)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s, %d\n", p, n)

	file, err := os.Create("writeAt.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.WriteString("Golang中文社区——这里是多余")
	n, err = file.WriteAt([]byte("Go语言中文网"), 24)
	if err != nil {
		panic(err)
	}

	fmt.Println(n)

	reader.WriteTo(os.Stdout)

	reader.Seek(-6, io.SeekEnd)
	r, _, _ := reader.ReadRune()
	fmt.Printf("\n%c\n", r)
}
