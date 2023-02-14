package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rnemeth90/httpbench"
	"github.com/spf13/pflag"
)

type config struct {
	url   string
	count int
}

var (
	url   string
	count int
)

func init() {
	pflag.StringVar(&url, "u", "", "url to test")
	pflag.IntVar(&count, "c", 4, "count of requests")
}

func usage() {
	fmt.Println(os.Args[0])

	fmt.Println("Usage:")
	fmt.Printf("  httpbench")

	fmt.Println("Options:")
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()

	c := config{
		url:   url,
		count: count,
	}

	if err := run(c, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}

func run(c config, w io.Writer) error {

	result, err := httpbench.MakeRequest(c.url)
	if err != nil {
		return err
	}

	fmt.Printf("%d", result.Status)

	return nil
}
