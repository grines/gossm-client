package completion

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/grines/ssmmmm-client/control/ssmaws"
)

func Upload(filename string, instID string) {

	//What file are we sending??
	out, err := CreateGzip(filename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//Chunk it
	chunks := Chunks(out, 400)
	cnt := len(chunks)
	msg := fmt.Sprintf("compressing file and splitting into %d chunks.", cnt)
	fmt.Println(msg)

	//Loop through and send chunksS
	var i int
	for _, v := range chunks {
		i++
		fmt.Println(i)
		ssmaws.UploadSendCommand(sess, "upload "+filename, instID, filename, v, i, cnt)
	}

}

func CreateGzip(filename string) (string, error) {
	fi, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(fi)
	w.Close()
	data := Base64Encode(string(b.Bytes()))
	return data, nil
}

func ReadGzFile(filename string) ([]byte, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	fz, err := gzip.NewReader(fi)
	if err != nil {
		return nil, err
	}
	defer fz.Close()

	s, err := ioutil.ReadAll(fz)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Chunks(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}
