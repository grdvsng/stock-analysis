package File

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func Read(path string) (b []byte, err error) {
	file, e := os.Open(path)

	if e != nil {
		return b, err
	}

	buffer := bytes.NewBufferString("")
	buf := make([]byte, 1024)

	for {
		n, e := file.Read(buf)

		if e == io.EOF || n == 0 {
			break
		} else if e != nil {
			err = e
			break
		}

		buffer.Write(buf[:n])
	}

	b = buffer.Bytes()
	return b, err
}

func ReadJson[T any](path string, instance T) (T, error) {
	bytes, e := Read(path)

	if e != nil {
		return instance, e
	}

	err := json.Unmarshal(bytes, &instance)

	return instance, err
}
