package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type TempResponse struct {
	Code    int                    `json:"code"`
	Name    string                 `json:"name"`
	Header  map[string]string      `json:"header"`
	Content map[string]interface{} `json:"content"`
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("index")

	body, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(body))

	ojson := make(map[string]interface{})
	json.Unmarshal(body, &ojson)

	tmpResp := &TempResponse{
		Code:    0,
		Name:    "whoami",
		Content: ojson,
	}

	tmpResp.Header = make(map[string]string)
	for k, v := range r.Header {
		tmpResp.Header[k] = v[0]
	}

	data, err := json.Marshal(&tmpResp)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, string(data))
}

func FooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("call foo")
	w.Write([]byte(fmt.Sprintf("doing foo by %s", os.Args[1])))
}

func BarHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("call bar")
	w.Write([]byte(fmt.Sprintf("doing bar by %s", os.Args[1])))
}

func CallerHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("do calling")
	r.ParseForm()
	surl := fmt.Sprintf("http://%s:%s/%s",
		r.Form.Get("ip"),
		r.Form.Get("port"),
		r.Form.Get("method"))
	resp, err := http.Get(surl)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("call http error :%s", err)))
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write([]byte(fmt.Sprintf("call http ok. header:%v body:%s",
		resp.Header, string(body))))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please input port")
		return
	}

	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/foo", FooHandler)
	http.HandleFunc("/bar", BarHandler)
	http.HandleFunc("/call", CallerHandler)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%s", os.Args[1]), nil)
		if err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second * 1)
	select {}
}
