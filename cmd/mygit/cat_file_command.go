package main

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func catFileCmd(args []string) (err error) {
	// Assuming that args is ["cat-file", "-p", "hash"], just like os.Args

	if len(args) < 3 || args[1] != "-p" {
		fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p <blob_hash>\n")
		return fmt.Errorf("bad usage")
	}

	blobSha := args[2]

	if len(blobSha) != 2*sha1.Size {
		return fmt.Errorf("not a valid object name: %v", blobSha)
	}

	path := filepath.Join(".git", "objects", blobSha[:2], blobSha[2:])

	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("not a valid object name: %v", blobSha)
	}
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	defer func() {
		e := file.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close file: %w", e)
		}
	}()

	return catFile(file)
}

func catFile(r io.Reader) (err error) {
	zr, err := zlib.NewReader(r)
	if err != nil {
		return fmt.Errorf("new zlib reader: %w", err)
	}

	defer func() {
		e := zr.Close()
		if err == nil && e != nil {
			err = fmt.Errorf("close zlib reader: %w", e)
		}
	}()

	err = parseObject(zr)
	if err != nil {
		return fmt.Errorf("parse object: %w", err)
	}

	return nil
}

func parseObject(r io.Reader) (err error) {
	br := bufio.NewReader(r)

	typ, err := br.ReadString(' ')
	if err != nil {
		return err
	}

	typ = typ[:len(typ)-1] // cut ' '

	if typ != "blob" {
		return fmt.Errorf("unsupported type: %v", typ)
	}

	sizeStr, err := br.ReadString('\000')
	if err != nil {
		return err
	}

	sizeStr = sizeStr[:len(sizeStr)-1] // cut '\000'

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("parse size: %w", err)
	}

	_, err = io.CopyN(os.Stdout, br, size)
	if err != nil {
		return err
	}

	return nil
}
