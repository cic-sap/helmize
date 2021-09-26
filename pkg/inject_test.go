package pkg

import (
	"os"
	"testing"
)

func TestInject(t *testing.T) {
	os.Setenv("HELM_NAMESPACE", "default")

	err := Inject("demo", "testdata/charts")
	if err != nil {
		t.Fatal(err)
	}

}
