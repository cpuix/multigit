package stdio

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// GetInput reads a line from the standard input and returns it.
func GetInput(l string) string {

	r, _ := io.Pipe()
	scanner := bufio.NewScanner(r)
	fmt.Fprintln(os.Stdout, l)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("scanner.Text():", scanner.Text())
	text := scanner.Text()
	if len(text) == 0 {
		log.Fatal("empty input")
	}

	fmt.Printf("You entered: %s\n", text)
	return text
}
