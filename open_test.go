// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package porteiro

import (
	"errors"
	"io"
	"reflect"
	"testing"
)

func TestRegisterAndOpen(t *testing.T) {
	var recorder callRecorder
	fn1 := makeFakeFn("1", nil, &recorder)
	fn2 := makeFakeFn("2", nil, &recorder)
	fn3 := makeFakeFn("3", nil, &recorder)
	fn4 := makeFakeFn("4", nil, &recorder)
	var opener *Opener
	opener = opener.Register("http", fn1).Register("ftp", fn2)
	opener = opener.Register("s3", fn3).Register("ftp", fn4)
	_, err := opener.Open("http://something-nice")
	if err != nil {
		t.Error(err)
	}
	_, err = opener.Open("ftp://hello-its-me")
	if err != nil {
		t.Error(err)
	}
	_, err = opener.Open("s3://some-bucket/object.txt")
	if err != nil {
		t.Error(err)
	}
	expectedCalls := []fakeFnCall{
		{id: "1", uri: "http://something-nice"},
		{id: "4", uri: "ftp://hello-its-me"},
		{id: "3", uri: "s3://some-bucket/object.txt"},
	}
	if !reflect.DeepEqual(recorder.calls, expectedCalls) {
		t.Errorf("wrong calls made\nwant %#v\ngot  %#v", expectedCalls, recorder.calls)
	}
}

func TestOpenUnkownScheme(t *testing.T) {
	var opener Opener
	rc, err := opener.Open("http://something-funny")
	if err == nil {
		t.Fatal("unexpected <nil> error")
	}
	expectedMsg := `can't open "http://something-funny": unknown scheme "http"`
	if msg := err.Error(); msg != expectedMsg {
		t.Errorf("wrong error message\nwant %q\ngot  %q", expectedMsg, err)
	}
	if rc != nil {
		t.Errorf("unexpected non-nil ReadCloser: %#v", rc)
	}
}

func TestFailureOnOpen(t *testing.T) {
	prepErr := errors.New("something went wrong")
	fn := makeFakeFn("1", prepErr, &callRecorder{})
	var opener *Opener
	opener = opener.Register("http", fn)
	rc, err := opener.Open("http://whatever")
	if err != prepErr {
		t.Errorf("wrong error returned\nwant %#v\ngot  %#v", prepErr, err)
	}
	if rc != nil {
		t.Errorf("unexpected non-nil ReadCloser: %#v", rc)
	}
}

type fakeFnCall struct {
	id  string
	uri string
}

type callRecorder struct {
	calls []fakeFnCall
}

func makeFakeFn(id string, err error, callRecorder *callRecorder) OpenFunc {
	return func(uri string) (io.ReadCloser, error) {
		callRecorder.calls = append(callRecorder.calls, fakeFnCall{id: id, uri: uri})
		return nil, err
	}
}
