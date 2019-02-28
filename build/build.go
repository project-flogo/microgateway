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
	case "clean":
		clean()
	case "test":
		test()
	case "generate":
		generate()
	case "help":
		fmt.Println(" clean - clean up")
		fmt.Println(" test - run full test")
	default:
		fmt.Println("[USAGE]: go run build.go [target]")
	}
}

func test() {
	command("go", "test", "...")
}


func clean(){

}


func generate(){
	command("echo", "echoing command in go")
}

func command(name  string, arg ...string) {
	output, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
	fmt.Println(string(output))
}