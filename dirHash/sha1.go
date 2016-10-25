package dirHash

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		go walkDir(dir, &wait, hashStr)
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

//读取需要被忽略的文件，由于时间关系没有实现通配符忽略文件
func getIgnoreFiles(dir string) map[string]bool {
	ignorefile := filepath.Join(dir, ".sha1Ignore")
	f, err := os.Open(ignorefile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		panic("read .sha1Ignore file error")
	}
	defer f.Close()

	rb := bufio.NewReader(f)
	ignores := make(map[string]bool)
	ignores[".sha1Ignore"] = true
	for {
		line, _, err := rb.ReadLine()
		if err != nil || io.EOF == err {
			if err == io.EOF {
				break
			}
			panic("read lines of .sha1Ignore file error")
		}
		ignores[string(line)] = true
	}
	return ignores
}

func walkDir(dir string, wait *sync.WaitGroup, hashStr chan<- string) {
	defer wait.Done()

	ignores := getIgnoreFiles(dir)
	for _, entry := range dirents(dir) {
		//忽略文件
		name := entry.Name()
		if ignores != nil && ignores[name] {
			continue
		}

		if entry.IsDir() {
			wait.Add(1)
			subdir := filepath.Join(dir, name)
			go walkDir(subdir, wait, hashStr)
		} else {
			fileFullName := filepath.Join(dir, name)
			fileSize := entry.Size()
			fileHash := getHash(fileFullName)
			hashStr <- fmt.Sprintf("%s, %x, %d\n", fileFullName, fileHash, fileSize)
		}
	}
}

// 用channel模拟信号量，控制goroutine的数量为15，
//防止文件打开过多而使用系统崩溃
var sema = make(chan struct{}, 15)

//读取文件或目录的基本信息
func dirents(dir string) []os.FileInfo {
	//读取信息量
	sema <- struct{}{}
	//使用完后，释放信号量
	defer func() { <-sema }()

	f, err := os.Open(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open error: %v\n", err)
		return nil
	}
	defer f.Close()

	entries, err := f.Readdir(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Readdir error: %v\n", err)
	}
	return entries
}
