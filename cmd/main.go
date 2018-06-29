package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	obfuscation "github.com/cszichao/image-obfuscation"
)

var (
	src      = flag.String("s", "", "image or folder to be obfuscated")
	parallel = flag.Int("t", 10, "max runing threads")
)

func main() {
	if _, err := exec.LookPath("pngquant"); err != nil {
		fmt.Fprintln(os.Stderr, "pngquant not installed")
		os.Exit(1)
	}
	if _, err := exec.LookPath("jpegoptim"); err != nil {
		fmt.Fprintln(os.Stderr, "jpegoptim not installed")
		os.Exit(1)
	}

	flag.Parse()
	fi, err := os.Stat(*src)
	if err != nil {
		fmt.Println("fail to access provided source " + *src + " with error " + err.Error())
		os.Exit(1)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		imageFiles := []string{}
		err := filepath.Walk(*src,
			func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext == ".png" ||
					ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
					imageFiles = append(imageFiles, path)
				}
				return nil
			})
		if err != nil {
			fmt.Println("fail to find image files in folder with error ", err)
		}
		processImages(imageFiles, *parallel)
	case mode.IsRegular():
		rand.Seed(time.Now().UTC().UnixNano())
		if strings.ToLower(path.Ext(*src)) == ".png" {
			exec.Command("pngquant", "--strip", "--skip-if-larger", "--ext=.png", "--force", *src).Run()
		}
		imgType, err := obfuscation.Obfuscate(*src, *src)
		if err != nil {
			println("fail to process file", *src)
			os.Exit(1)
		}
		if fi.Size() > 64*1024 { // only shrink images over 64k
			if imgType == obfuscation.ImageTypePNG {
				exec.Command("pngquant", "--strip", "--skip-if-larger",
					"--ext=.png", "--force", "--quality", "0-"+strconv.Itoa(90+rand.Intn(10)), *src).Run()
			} else if imgType == obfuscation.ImageTypeJPG {
				exec.Command("jpegoptim", "--strip", "-m", strconv.Itoa(80+rand.Intn(20)), *src).Run()
			}
		}
	}

}

func processImages(images []string, threads int) {
	processQueue := make(chan string, threads)
	go func() {
		for _, image := range images {
			processQueue <- image
		}
		close(processQueue)
	}()
	wg := sync.WaitGroup{}
	wg.Add(threads)
	for t := 0; t < threads; t++ {
		go func() {
			for image := range processQueue {
				// use process exec since FFTW isn't thread safe
				// do NOT obfuscation.Obfuscate(image, image)
				fmt.Println(exec.Command(os.Args[0], "-s", image, image).Run())
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
