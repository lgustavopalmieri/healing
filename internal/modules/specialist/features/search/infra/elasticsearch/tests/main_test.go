package tests

import (
	"fmt"
	"testing"

	elasticsearchtest "github.com/lgustavopalmieri/healing-specialist/internal/commom/tests/elasticsearch"
)

var testHelper *elasticsearchtest.TestHelper

func TestMain(m *testing.M) {
	fmt.Println("TestMain is being executed")
	testHelper = elasticsearchtest.NewTestHelper()
	if testHelper == nil {
		panic("testHelper is nil after NewTestHelper()")
	}
	testHelper.RunTestMain(m)
}
