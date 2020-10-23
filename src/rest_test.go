package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunRestServer(t *testing.T) {
	go InitServer()
	time.Sleep(2 * time.Second)
	longText := "VeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongText" +
		"VeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongTextVeryLongText" +
		"VeryLongTextVeryLongText"

	resp, err := http.Post("http://"+resolveHostIp(":8081")+"/objects/"+longText, "", nil)
	if err != nil {
		t.Error("Could not get POST response")
		return
	}

	if resp.StatusCode != http.StatusBadRequest {
		fmt.Println(resp.StatusCode)
		t.Error("Should have returned a bad request error")
		return
	}

	text := "This is a test"
	resp, err = http.Post("http://"+resolveHostIp(":8081")+"/objects/"+text, "", nil)
	if err != nil {
		t.Error("Could not get POST response")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Could not decode body")
		return
	}
	stringBody := string(body)
	res := strings.Split(stringBody, "\n")[3]

	textShaByte := sha1.Sum([]byte(text))
	if res != hex.EncodeToString(textShaByte[:]) {
		t.Error("Incorrect Sha1 value returned")
		return
	}

	resp, err = http.Get("http://" + resolveHostIp(":8081") + "/objects/" + res)
	if err != nil {
		t.Error("Could not get GET response")
		return
	}
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Could not decode body")
		return
	}
	stringBody = string(body)
	res = strings.Split(stringBody, "\n")[3]
	if res != text {
		t.Error("Incorrect data returned")
	}
}
