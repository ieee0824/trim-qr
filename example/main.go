package main

import (
	"os"
	"log"
	"image/jpeg"
	"github.com/ieee0824/trim-qr"
	"bytes"
	"io/ioutil"
)

func main() {
	f, err := os.Open("qr_1024.jpg")
	if err != nil {
		log.Fatalln(err)
	}

	img, err := jpeg.Decode(f)
	if err != nil {
		log.Fatalln(err)
	}

	gray,err := tqr.Tqr(img)
	if err != nil {
		log.Fatalln(err)
	}


	buf := new(bytes.Buffer)

	if err := jpeg.Encode(buf, gray, nil); err != nil {
		log.Fatalln(err)
	}

	if err := ioutil.WriteFile("qr.jpeg", buf.Bytes(), 0644); err != nil{
		log.Fatalln(err)
	}
}
