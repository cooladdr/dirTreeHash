package dirHash

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func ComputingHash(dirs []string) (hashFile string) {
	//用来返回文件hash结果
	hashStr := make(chan string)

	//用来行等待所有goroutine完成
	var wait sync.WaitGroup

	for _, dir := range dirs {
		wait.Add(1)
		go walkDir2(dir, &wait, hashStr)
	}

	//等待所有goroutine完成，然后关闭channel
	go func() {
		wait.Wait()
		close(hashStr)
	}()

	//读取每个goroutine返回的hash结果
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	outFilePath := "./hashResult/hashout" + timeStamp + ".txt"
	outFile, err := os.Create(outFilePath)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	for str := range hashStr {
		outFile.WriteString(str)
	}

	return outFilePath
}

func getHash(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Open:%s error", file)
		panic(err)
	}
	defer f.Close()

	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		panic(err)
	}

	return h.Sum(nil)
}

type ignoreFiles struct {
	ignores map[string]bool
}

//读取需要被忽略的文件，由于时间关系没有实现通配符忽略文件
func (i *ignoreFiles) readIgnoreFiles(dir string) {
	ignorefile := filepath.Join(dir, ".sha1Ignore")
	f, err := os.Open(ignorefile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		panic("read .sha1Ignore file error")
	}
	defer f.Close()

	rb := bufio.NewReader(f)
	i.ignores = make(map[string]bool)
	i.ignores[".sha1Ignore"] = true
	for {
		line, _, err := rb.ReadLine()
		if err != nil || io.EOF == err {
			if err == io.EOF {
				break
			}
			panic("read lines of .sha1Ignore file error")
		}
		i.ignores[string(line)] = true
	}
}

func (i *ignoreFiles) canIgnore(filePath string) bool {
	if i.ignores == nil {
		return false
	}

	for name, state := range i.ignores {
		if state && strings.HasPrefix(strings.Replace(filePath, "\\", "/", -1), name) {
			return true
		}
	}

	return false
}

var sema2 = make(chan struct{}, 30)

func walkDir2(dir string, wait *sync.WaitGroup, hashStr chan<- string) {
	defer wait.Done()

	ignore := new(ignoreFiles)
	ignore.readIgnoreFiles(dir)

	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		if err != nil {
			fmt.Println(err)
			return nil
		}

		if ignore.canIgnore(path) {
			return nil
		}

		wait.Add(1)
		go func(goPath string, fileInfo os.FileInfo, goHashStr chan<- string) {
			defer wait.Done()
			sema2 <- struct{}{}
			defer func() { <-sema2 }()

			fileSize := fileInfo.Size()
			fileHash := getHash(goPath)
			goHashStr <- fmt.Sprintf("%s, %x, %d\n", goPath, fileHash, fileSize)
		}(path, fi, hashStr)

		return nil
	})
}
