package main

import (
	"testing"
	"bytes"
)

func TestStorage(t *testing.T) {
	st := NewStorage()
	st.Store("test1", []byte("asdf"), maxExpire, false)
	st.Store("test2", []byte("qwerty"), maxExpire, false)
	st.Store("test2", []byte("qwerty"), maxExpire, false)
	v, ok := st.Load("test1")
	if !bytes.Equal(v, []byte("asdf")) {
		t.Errorf("st.Load() = %s; want asdf", string(v))
	}
	if ok == false {
		t.Errorf("st.Load() = _, false; want true")
	}
	st.Delete("test1", false)
	v2, ok2 := st.Load("test1")
	if v2 != nil {
		t.Errorf("st.Load(\"test1\") = %d; want nil", v)
	}
	if ok2 == true {
		t.Errorf("st.Load(\"test1\") = true; want false")
	}
    st.RefreshDataPeriodically("test2", 10)
    ticker, ok3 := st.refreshStorage["test2"]
	if !ok3 {
		t.Errorf("expected refresh storage to include test2")
	}
    go st.StopDataRefresh("test2")
    <- ticker.forgetCh 
    _, ok4 := st.refreshStorage["test2"]
	if ok4 == true {
		t.Errorf("expected refresh storage to not include test2, since it is forgotten")
	}
    
}
