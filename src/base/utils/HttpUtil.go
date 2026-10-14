package utils

import (
	"gamekit/src/base/logger"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"
)

type ResponseRecorder struct {
	header http.Header
	body   []byte
	status int
}

func NewResponseRecorder() *ResponseRecorder {
	return &ResponseRecorder{
		header: http.Header{},
	}
}

func (rec *ResponseRecorder) Header() http.Header {
	return rec.header
}

func (rec *ResponseRecorder) Write(data []byte) (int, error) {
	rec.body = append(rec.body, data...)
	return len(data), nil
}

func (rec *ResponseRecorder) WriteHeader(statusCode int) {
	rec.status = statusCode
}

func (rec *ResponseRecorder) StatusCode() int {
	return rec.status
}

func (rec *ResponseRecorder) Body() []byte {
	return rec.body
}

/**
* @author: Tavin
* @date: 2024/4/24
 */

// DoPbHttpRequest 发起Protobuf请求
func DoPbHttpRequest[T proto.Message](url string, msg proto.Message) (T, error) {
	var t T
	var request *http.Request
	requestBody := &bytes.Buffer{}
	if nil != msg {
		requestBytes, err := proto.Marshal(msg)
		if nil != err {
			return t, err
		}
		requestBody = bytes.NewBuffer(requestBytes)
	}
	request, err := http.NewRequest("POST", url, requestBody)
	if nil != err {
		return t, err
	}
	request.Header.Set("Content-Type", "application/x-protobuf")

	//	requestDump, err := httputil.DumpRequestOut(request, true)
	//	if err == nil {
	//		logger.Log.Debug("appstore request: %s", string(requestDump))
	//	}

	resp, err := Client.Do(request)
	defer func() {
		if nil != resp {
			resp.Body.Close()
		}
	}()
	if nil != err {
		return t, err
	}
	if nil == resp {
		return t, errors.New("DoPbHttpRequest | resp is nil")
	}
	if resp.StatusCode != http.StatusOK {
		return t, errors.New(resp.Status)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if nil != err {
		return t, err
	}
	dist := reflect.New(reflect.TypeOf(t).Elem()).Interface().(proto.Message)
	err = proto.Unmarshal(responseBody, dist)
	if nil != err {
		return t, err
	}
	return dist.(T), nil
}

func GenSum(params map[string]interface{}, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var strParam string

	for _, k := range keys {
		value := GetMapString(params, k)
		if len(strParam) <= 0 {
			strParam = fmt.Sprintf("%s=%s", k, value)
		} else {
			strParam += fmt.Sprintf("&%s=%s", k, value)
		}
	}
	strParam += key

	return Md5Sum(strParam)
}

// 新平台请求
func DoHttpPost(url string, params map[string]interface{}, secretKey string, response interface{}) error {
	now := time.Now()

	params["timestamp"] = now.Unix()
	sign := GenSum(params, secretKey)
	params["sign"] = sign

	data, err := json.Marshal(params)
	if err != nil {
		logger.Log.Error("doHttp failed. url: %s, params: %+v, error: %s", url, params, err.Error())
		return err
	}

	resp, err := Client.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		logger.Log.Error("doHttp failed. url: %s, params: %+v, err: %s", url, params, err.Error())
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.Error("doHttp failed. url: %s, params: %+v, code: %d", url, params, resp.StatusCode)
		return fmt.Errorf("http code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("http read response failed. apiUrl: %s, params: %+v, err: %s", url, params, err.Error())
		return err
	}

	logger.Log.Error("url: %s, req: %+v, resp: %s", url, params, string(body))

	err = json.Unmarshal(body, response)
	if err != nil {
		logger.Log.Error("unmarshal response failed. url: %s, body: %s, err: %s", url, string(body), err.Error())
		return err
	}

	return nil
}
