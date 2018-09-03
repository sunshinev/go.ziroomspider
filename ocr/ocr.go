package ocr

import (
	"fmt"
	"net/http"
	"os"
	"github.com/otiai10/gosseract"
	"io/ioutil"
	"io"
	"bytes"
)

func Parse(imageUrl string)(string) {

	f, _ := os.Create("ziroomspider.png")
	defer f.Close()

	resp, _ := http.Get(imageUrl)
	defer  resp.Body.Close()

	pic, _ := ioutil.ReadAll(resp.Body)
	io.Copy(f, bytes.NewReader(pic))

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("./ziroomspider.png")
	text, e := client.Text()

	if e != nil {
		fmt.Println("error")
	}

	return text
}
