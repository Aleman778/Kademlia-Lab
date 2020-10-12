package main

import (
	"bytes"
	"testing"
)

func TestStorage(t *testing.T) {
	st := NewStorage()
	st.Store("test1", []byte("asdf"))
	st.Store("test2", []byte("qwerty"))
	st.Store("test2", []byte("qwerty"))
	v, ok := st.Load("test1")
	if !bytes.Equal(v, []byte("asdf")) {
		t.Errorf("st.Load() = %s; want asdf", string(v))
	}
	if ok == false {
		t.Errorf("st.Load() = _, false; want true")
	}
	st.Delete("test1")
	v2, ok2 := st.Load("test1")
	if v2 != nil {
		t.Errorf("st.Load(\"test1\") = %d; want nil", v)
	}
	if ok2 == true {
		t.Errorf("st.Load(\"test1\") = true; want false")
	}
}
