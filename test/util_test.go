package test

import (
	"github.com/Azer0s/quacktors/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGoid(t *testing.T) {
	a := assert.New(t)

	a.Equal(util.GetGoid(), util.GetGoid())
}

func TestParseRemoteAddress(t *testing.T) {
	as := assert.New(t)

	a := "foo@localhost:5521"
	b := "foo@test.com:5521"
	c := "foo@:5521"
	d := "foo@127.0.0.1:5521"
	e := "foo@[1762:0:0:0:0:B03:1:AF18]:5521"
	f := "foo@[64:ff9b::]:5521"
	g := "foo@[2001:db8:3:4::]:5521"

	system, addr, p, err := util.ParseAddress(a)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("localhost", addr)

	system, addr, p, err = util.ParseAddress(b)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("test.com", addr)

	system, addr, p, err = util.ParseAddress(c)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("", addr)

	system, addr, p, err = util.ParseAddress(d)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("127.0.0.1", addr)

	system, addr, p, err = util.ParseAddress(e)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("1762:0:0:0:0:B03:1:AF18", addr)

	system, addr, p, err = util.ParseAddress(f)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("64:ff9b::", addr)

	system, addr, p, err = util.ParseAddress(g)
	if err != nil {
		t.Fatal(err)
	}
	as.Equal(5521, p)
	as.Equal("foo", system)
	as.Equal("2001:db8:3:4::", addr)

}
