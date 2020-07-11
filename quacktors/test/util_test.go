package test

import (
	"github.com/Azer0s/Quacktors/quacktors/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGoid(t *testing.T) {
	a := assert.New(t)

	a.Equal(util.GetGoid(), util.GetGoid())
}
