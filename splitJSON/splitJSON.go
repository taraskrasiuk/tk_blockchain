package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
)

func SplitJSON(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	d := data[:]

	// st, en := bytes.IndexByte(data, '['), bytes.IndexByte(data, ']')
	// if st == -1 || en == -1 {
	// 	return 0, []byte{}, errors.New("expected json to be around with '[ ... ]'")
	// }

	// d := data[st:en]
	// curly
	findClose := func(data []byte) int {
		d := data[:]
		for idx := bytes.IndexByte(d, '}'); idx != -1; {
			if len(d) >= idx+1 {
				nextCh := d[idx+1]
				if nextCh == ',' {
					return idx
				}
				d = d[idx:]
			}
			break
		}
		return -1
	}
	stc, enc, encE := bytes.IndexByte(d, '{'), bytes.IndexByte(d, '}'), findClose(d)
	fmt.Println(stc, enc)
	if stc > 0 {
		if enc == -1 && encE == -1 {
			fmt.Println(string(d[bytes.IndexByte(d, '}')]))
			if d[enc+1] != ']' {
				return 0, []byte{}, errors.New("expected curly to be closed")
			}
		}
		if encE == -1 {
			dd := d[stc : enc+1]
			return len(dd), dd, nil
		}
		dd := d[stc : encE+1]
		return len(dd), dd, nil
	}

	if atEOF {
		if len(data) > 0 && data[len(data)-1] == '\r' {
			return len(data), data[:len(data)-1], nil
		}
		return len(data), data, nil
	}
	return 0, []byte{}, nil
}

func main() {
	input := `
		[
			{
				"name":  "taras",
				"secondname": "krasiuk"
			},
			{
			"name":  "inna",
			"secondname": "ra"
			}
		]
	`

	sc := bufio.NewScanner(bytes.NewReader([]byte(input)))
	sc.Split(SplitJSON)

	for sc.Scan() {
		fmt.Println(sc.Text())
	}
	if sc.Err() != nil {
		log.Fatal(sc.Err())
	}
}
