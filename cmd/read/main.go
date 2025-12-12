package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("./data/test.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := []byte{}
	for {
		fmt.Printf("start: len(buf) = %d | cap(buf) = %d\n", len(buf), cap(buf))
		b := make([]byte, 15)
		fmt.Printf("bread:   len(b) = %d |   cap(b) = %d\n", len(b), cap(b))
		_, err := file.Read(b)
		fmt.Printf("aread:   len(b) = %d |   cap(b) = %d\n", len(b), cap(b))
		buf = append(buf, b...)
		fmt.Printf("aread: len(buf) = %d | cap(buf) = %d\n", len(buf), cap(buf))

		if err == io.EOF {
			break
		}
	}

	fmt.Printf("  end: len(buf) = %d | cap(buf) = %d\n", len(buf), cap(buf))
	fmt.Printf("%s\n", buf)
}
