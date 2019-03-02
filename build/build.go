// +build ignore

package main

import (
	"flag"
	"fmt"
	//"io/ioutil"
	"os"
	"os/exec"
	//"path/filepath"
	//"reflect"
	//"runtime"
	//"strings"
)

func main() {
	flag.Parse()

	args := flag.Args()
	target := ""
	if len(args) > 0 {
		target = args[0]
	}

	switch target {
	case "build":
		build()
	case "clean":
		clean()
	case "test":
		test()
	case "test-short":
		testshort()
	case "generate":
		generate()
	case "help":
		fmt.Println(" clean - clean up")
		fmt.Println(" test - run full test")
		fmt.Println(" test-short - skip integration tests")
		fmt.Println(" generate - generates code as per directives")
	default:
		fmt.Println("[USAGE]: go run build.go [target]")
	}
}

func build(){
	generate()
	command("go", "build")
}

func test() {
	build()
	command("go", "clean", "-testcache")
	command("go", "test","-p","1","./...")
}

func testshort() {
	build()
	command("go", "clean", "-testcache")
	command("go", "test","-p","1","-short","./...")
}


func clean(){
	command("go", "clean")
	command("go", "clean", "-testcache")
}


func generate(){
	command("go", "generate", "./...")
}

func command(name  string, arg ...string) {
	output, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	fmt.Println(string(output))
}