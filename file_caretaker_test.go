package orchestrator

import "testing"

func TestWriteAndRead(t *testing.T) {
	fc, _ := NewFileCareTacker("sample")
	defer fc.shutdown()

	fc.persist("id", "some sam ple data BB")
	result, err := fc.get("id")

	if result == "" || err != nil {
		t.Fail()
	}
}
