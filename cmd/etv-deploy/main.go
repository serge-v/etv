package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

var verbose = flag.Bool("verbose", false, "print commands")

func main() {
	flag.Parse()
	if *verbose {
		log.Printf("cd ../etv")
	}
	if err := os.Chdir("../etv"); err != nil {
		log.Fatal(err)
	}
	generate()
	build()
	deploy()
}

func getVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--long", "--dirty")
	if *verbose {
		log.Printf("%v\n", cmd.Args)
	}
	buf, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(buf))
}

func generate() {
	cmd := exec.Command("go", "test")
	if *verbose {
		log.Printf("%v\n", cmd.Args)
	}
	buf, err := cmd.CombinedOutput()
	fmt.Println(string(buf))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}

func build() {
	ver := getVersion()
	cmd := exec.Command("go", "build", "-ldflags", "-X main.version="+ver, "-o", "etv")
	if *verbose {
		log.Printf("%v\n", cmd.Args)
	}
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=arm", "GOARM=5")
	buf, err := cmd.CombinedOutput()
	fmt.Println(string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func run(program string, args ...string) {
	cmd := exec.Command(program, args...)
	if *verbose {
		log.Printf("%v\n", cmd.Args)
	}
	buf, err := cmd.CombinedOutput()
	fmt.Println(string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func runterm(program string, args ...string) {
	cmd := exec.Command(program, args...)
	if *verbose {
		log.Printf("%v\n", cmd.Args)
	}
	cmd.Stdin = os.Stdin
	buf, err := cmd.CombinedOutput()
	fmt.Println(string(buf))
	if err != nil {
		log.Fatal(err)
	}
}

func deploy() {
	fname := os.Getenv("HOME") + "/.config/etv/deploy.txt"
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	dst := strings.TrimSpace(string(buf))
	run("scp", "etv", dst+":/tmp/etv.new")
	runterm("ssh", "-t", dst, "su", "-c", `"./writeenable.sh; cp /tmp/etv.new etv; ./etv -version"`)
}
