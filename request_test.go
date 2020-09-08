/**
* Created by GoLand.
* User: luffy
* Date: 2019-05-13
* Time: 19:46
 */
package greq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func server(handler http.HandlerFunc) *httptest.Server {
	if handler == nil {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello luffy !!"))
		}))
	}
	return httptest.NewServer(handler)
}

func TestParamQuery(t *testing.T) {
	params := url.Values{}
	params.Set("foo", "bar")
	params.Set("name", "中文")
	handler := func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for key, value := range params {
			if v := query.Get(key); v != value[0] {
				t.Errorf("got %s, want %s", v, value[0])
			}
		}
		w.Write([]byte(query.Encode()))
	}
	ts := server(handler)
	req := NewRequest("get", ts.URL)
	req.SetParams(params)
	resp := req.Exec()
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}

	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
}

func TestPostForm(t *testing.T) {
	params := url.Values{}
	params.Set("foo", "bar")
	params.Set("hello", "world")
	handler := func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			t.Errorf("r.ParseForm error, got = %s", err.Error())
		}
		for key, value := range params {
			if v := r.FormValue(key); v != value[0] {
				t.Errorf("form value got = %s, want = %s", v, value[0])
			}
		}
		w.Write([]byte(r.PostForm.Encode()))
	}
	ts := server(handler)
	req := NewRequest("post", ts.URL)
	req.SetParams(params)
	resp := req.Exec()
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}

	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
}

func TestParamBody(t *testing.T) {
	reqBody := "success"
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll error, err = %s", err.Error())
		}
		if string(body) != reqBody {
			t.Errorf("got body %s, want reqBody %s", body, reqBody)
		}
		w.Write([]byte(body))
	}
	ts := server(handler)
	params := url.Values{}
	req := NewRequest("post", ts.URL)
	req.SetBody(strings.NewReader(reqBody))
	params.Set("foo", "bar")
	req.SetParams(params)
	resp := req.Exec()
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}

	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
	bytess, err := resp.ToBytes()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", bytess)
}

func TestJSONBody(t *testing.T) {
	type content struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	reqBody := content{Code: "0", Msg: "success"}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Errorf("json.Marshal reqBody error, %s", err.Error())
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("ioutil.ReadAll error %s", err.Error())
		}
		if string(body) != string(bodyBytes) {
			t.Errorf("got body %s, want body %s", body, bodyBytes)
		}
		w.Write(body)
	}
	ts := server(handler)
	req := NewRequest("post", ts.URL)
	req.SetContentType(TypeJSON)
	req.SetBody(bytes.NewReader(bodyBytes))
	resp := req.Exec()
	var c content
	err = resp.ToJSON(&c)
	if err != nil {
		t.Errorf("resp.ToJSON error, err = %s", err.Error())
	}
	if c != reqBody {
		t.Errorf("got body = %s, want body = %s", c, reqBody)
	}
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}
	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)

	req = NewRequest("post", ts.URL)
	req.SetBodyJSON(reqBody)
	resp = req.Exec()
	var cc content
	err = resp.ToJSON(&cc)
	if err != nil {
		t.Errorf("resp.ToJSON error, err = %s", err.Error())
	}
	if cc != reqBody {
		t.Errorf("got body = %s, want body = %s", c, reqBody)
	}
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}
	dumpRequest, _ = resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ = resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err = resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
}

func TestCookie(t *testing.T) {
	expiration := time.Now().Add(time.Minute * 5)
	cookie := &http.Cookie{Name: "Foo", Value: "bar", Expires: expiration}
	handler := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("Foo")
		if err != nil {
			t.Errorf("r.Cookie error, err = %s", err.Error())
		}
		if c.Value != cookie.Value {
			t.Errorf("cookie got = %#v, want = %#v", c, cookie)
		}
	}
	ts := server(handler)
	req := NewRequest("post", ts.URL)
	req.AddCookie(cookie)
	req.SetParam("foo", "bar")
	resp := req.Exec()
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}

	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
}

//func TestFile(t *testing.T) {
//	handler := func(w http.ResponseWriter, r *http.Request) {
//		err := r.ParseMultipartForm(32 << 20)
//		if err != nil {
//			t.Errorf("ParseMultipartForm error, err = %s", err.Error())
//		}
//		file, h, err := r.FormFile("uploadfile")
//		if err != nil {
//			t.Errorf("Form file error, err = %s", err.Error())
//		}
//		defer file.Close()
//		t.Logf("upload file: %+v \n", h.Filename)
//		t.Logf("file size: %+v \n", h.Size)
//		t.Logf("MIME header: %+v \n", h.Header)
//
//		tempFile, err := ioutil.TempFile("./", "*.jpg")
//		if err != nil {
//			t.Errorf("create tempfile error, err = %s", err.Error())
//		}
//		defer tempFile.Close()
//		fileBytes, err := ioutil.ReadAll(file)
//		if err != nil {
//			t.Errorf("ReadAll file error, err = %s", err.Error())
//		}
//		len, err := tempFile.Write(fileBytes)
//		if err != nil {
//			t.Errorf("tempFile.write error, err = %s", err.Error())
//		}
//		t.Logf("success upload file, len = %d ", len)
//		w.Write([]byte("success"))
//	}
//	ts := server(handler)
//	req := NewRequest("post", ts.URL)
//	req.SetFile("uploadfile", "0326.jpg", "0326.jpg")
//	req.SetParam("tt", "0326")
//	resp := req.Exec()
//	if err := resp.Error(); err != nil {
//		t.Errorf("req.exec error err= %s", err.Error())
//	}
//	if resp.StatusCode() != 200 {
//		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
//	}
//
//	dumpRequest, _ := resp.DumpRequest(true)
//	fmt.Printf("dumpRequest %s \n", dumpRequest)
//	dumpResponse, _ := resp.DumpResponse(true)
//	fmt.Printf("dumpResponse %s \n", dumpResponse)
//	body, err := resp.ToString()
//	if err != nil {
//		t.Errorf("resp.ToString error, got = %s \n", err.Error())
//	}
//	fmt.Printf("body = %s \n", body)
//}
func TestProxy(t *testing.T) {
	req := NewRequest("get", "http://www.google.com")
	req.SetProxy("http://127.0.0.1:10080")
	req.SetTimeout(time.Second * 10)
	resp := req.Exec()
	if err := resp.Error(); err != nil {
		t.Errorf("req.exec error err= %s", err.Error())
		return
	}
	if resp.StatusCode() != 200 {
		t.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
	}

	dumpRequest, _ := resp.DumpRequest(true)
	fmt.Printf("dumpRequest %s \n", dumpRequest)
	dumpResponse, _ := resp.DumpResponse(true)
	fmt.Printf("dumpResponse %s \n", dumpResponse)
	body, err := resp.ToString()
	if err != nil {
		t.Errorf("resp.ToString error, got = %s \n", err.Error())
	}
	fmt.Printf("body = %s \n", body)
}

func Benchmark_withJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		type content struct {
			Code string `json:"code"`
			Msg  string `json:"msg"`
		}
		reqBody := content{Code: "0", Msg: "success"}
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			b.Errorf("json.Marshal reqBody error, %s", err.Error())
		}
		handler := func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				b.Errorf("ioutil.ReadAll error %s", err.Error())
			}
			if string(body) != string(bodyBytes) {
				b.Errorf("got body %s, want body %s", body, bodyBytes)
			}
			w.Write(body)
		}
		ts := server(handler)
		req := NewRequest("post", ts.URL)
		req.SetContentType(TypeJSON)
		req.SetBody(bytes.NewReader(bodyBytes))
		resp := req.Exec()
		var c content
		err = resp.ToJSON(&c)
		if err != nil {
			b.Errorf("resp.ToJSON error, err = %s", err.Error())
		}
		if c != reqBody {
			b.Errorf("got body = %s, want body = %s", c, reqBody)
		}
		if err := resp.Error(); err != nil {
			b.Errorf("req.exec error err= %s", err.Error())
		}
		if resp.StatusCode() != 200 {
			b.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
		}
		//dumpRequest, _ := resp.DumpRequest(true)
		//fmt.Printf("dumpRequest %s \n", dumpRequest)
		//dumpResponse, _ := resp.DumpResponse(true)
		//fmt.Printf("dumpResponse %s \n", dumpResponse)
		body, err := resp.ToString()
		if err != nil {
			b.Errorf("resp.ToString error, got = %s \n", err.Error())
		}
		fmt.Printf("body = %s \n", body)

		req = NewRequest("post", ts.URL)
		req.SetBodyJSON(reqBody)
		resp = req.Exec()
		var cc content
		err = resp.ToJSON(&cc)
		if err != nil {
			b.Errorf("resp.ToJSON error, err = %s", err.Error())
		}
		if cc != reqBody {
			b.Errorf("got body = %s, want body = %s", c, reqBody)
		}
		if err := resp.Error(); err != nil {
			b.Errorf("req.exec error err= %s", err.Error())
		}
		if resp.StatusCode() != 200 {
			b.Errorf("req.exec statuscode want = 200, got = %d", resp.StatusCode())
		}
		//dumpRequest, _ = resp.DumpRequest(true)
		//fmt.Printf("dumpRequest %s \n", dumpRequest)
		//dumpResponse, _ = resp.DumpResponse(true)
		//fmt.Printf("dumpResponse %s \n", dumpResponse)
		body, err = resp.ToString()
		if err != nil {
			b.Errorf("resp.ToString error, got = %s \n", err.Error())
		}
		fmt.Printf("body = %s \n", body)
	}
}
