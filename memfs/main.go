package main

import (
	"context"
	"io"
	"io/fs"
	"time"

	"github.com/Shopify/go-storage"
	"github.com/absfs/memfs"
	"github.com/liamg/memoryfs"
	. "github.com/stevegt/goadapt"
)

func liamg() {

	memfs := memoryfs.New()

	if err := memfs.MkdirAll("my/dir", 0o700); err != nil {
		panic(err)
	}

	dircount := 1000
	filecount := 1000
	count := dircount * filecount

	start := time.Now()
	for dir := 0; dir < dircount; dir++ {
		subdir := Spf("my/dir/subdir%d", dir)
		if err := memfs.MkdirAll(subdir, 0o700); err != nil {
			panic(err)
		}
		for file := 0; file < filecount; file++ {
			fn := Spf("%s/file%d.txt", subdir, file)
			err := memfs.WriteFile(fn, []byte("hello world"), 0o600)
			if err != nil {
				Pf("fn: %s\n", fn)
				panic(err)
			}
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	Pf("Elapsed time: %v\n", elapsed)
	Pf("write ops/sec: %v\n", float64(count)/elapsed.Seconds())

	start = time.Now()
	for dir := 0; dir < dircount; dir++ {
		subdir := Spf("my/dir/subdir%d", dir)
		for file := 0; file < filecount; file++ {
			fn := Spf("%s/file%d.txt", subdir, file)
			_, err := fs.ReadFile(memfs, fn)
			if err != nil {
				panic(err)
			}
		}
	}
	stop = time.Now()
	elapsed = stop.Sub(start)
	Pf("Elapsed time: %v\n", elapsed)
	Pf("read ops/sec: %v\n", float64(count)/elapsed.Seconds())

}

func shopifyfs() {
	mem := storage.NewMemoryFS()

	count := 1000000

	/*
		XXX unsupported
		if err := mem.Mkdir("foo/bar", 0o700); err != nil {
			panic(err)
		}
	*/

	for i := 0; i < count; i++ {
		fn := Spf("file%d.txt", i)
		fh, err := mem.Create(context.Background(), fn, nil)
		if err != nil {
			panic(err)
		}
		if _, err := io.WriteString(fh, "Hello World!"); err != nil {
			panic(err)
		}
		if err := fh.Close(); err != nil {
			panic(err)
		}
	}

	for i := 0; i < count; i++ {
		fn := Spf("file%d.txt", i)
		fh, err := mem.Open(context.Background(), fn, nil)
		if err != nil {
			panic(err)
		}
		_, err = io.ReadAll(fh)
		if err != nil {
			panic(err)
		}
		if err := fh.Close(); err != nil {
			panic(err)
		}
	}

}

func absfs() {

	fs, err := memfs.NewFS() // remember kids don't ignore errors
	Ck(err)

	dircount := 1000
	filecount := 1000
	count := dircount * filecount

	start := time.Now()
	for dir := 0; dir < dircount; dir++ {
		subdir := Spf("my/dir/subdir%d", dir)
		if err := fs.MkdirAll(subdir, 0o700); err != nil {
			panic(err)
		}
		for file := 0; file < filecount; file++ {
			fn := Spf("%s/file%d.txt", subdir, file)
			fh, err := fs.Create(fn)
			if err != nil {
				panic(err)
			}
			if _, err := fh.Write([]byte("hello world")); err != nil {
				panic(err)
			}
			if err := fh.Close(); err != nil {
				panic(err)
			}
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	Pf("Elapsed time: %v\n", elapsed)
	Pf("write ops/sec: %v\n", float64(count)/elapsed.Seconds())

	start = time.Now()
	for dir := 0; dir < dircount; dir++ {
		subdir := Spf("my/dir/subdir%d", dir)
		for file := 0; file < filecount; file++ {
			fn := Spf("%s/file%d.txt", subdir, file)
			fh, err := fs.Open(fn)
			if err != nil {
				panic(err)
			}
			_, err = io.ReadAll(fh)
			if err != nil {
				panic(err)
			}
			if err := fh.Close(); err != nil {
				panic(err)
			}
		}
	}
	stop = time.Now()
	elapsed = stop.Sub(start)
	Pf("Elapsed time: %v\n", elapsed)
	Pf("read ops/sec: %v\n", float64(count)/elapsed.Seconds())

	// fs.Remove("example.txt")

}

func main() {
	// liamg()
	// shopifyfs()
	absfs()
}
