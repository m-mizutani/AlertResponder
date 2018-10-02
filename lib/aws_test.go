package lib_test

import (
	"testing"

	"github.com/m-mizutani/AlertResponder/lib"
	"github.com/stretchr/testify/assert"
)

func TestLambdaArn(t *testing.T) {
	testArn := "arn:aws:lambda:ap-northeast-1:1234567890:function:mizutani-test"
	arn := lib.NewArn(testArn)

	assert.Equal(t, arn.Region(), "ap-northeast-1")
	assert.Equal(t, arn.FuncName(), "mizutani-test")
}
