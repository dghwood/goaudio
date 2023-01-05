package wav

import (
	"os"
	"testing"
)

func TestFromBytes(t *testing.T) {
	fileBytes, err := os.ReadFile("../samples/easy.wav")
	if err != nil {
		t.Error(err)
	}
	wavFile, err := FromBytes(fileBytes)
	if err != nil {
		t.Error(err)
	}
	if wavFile.Length != 1695744 {
		t.Error("parse Failed")
	}
}

func TestAppend(t *testing.T) {
	fileBytes, err := os.ReadFile("../samples/easy.wav")
	if err != nil {
		t.Error(err)
	}
	wavFile, err := FromBytes(fileBytes)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < 3; i++ {

		ifileBytes, err := os.ReadFile("../samples/easy.wav")
		if err != nil {
			t.Error(err)
		}
		iwavFile, err := FromBytes(ifileBytes)
		if err != nil {
			t.Error(err)
		}
		if wavFile.Append(&iwavFile) != nil {
			t.Error(err)
		}
	}
	fileBytes, err = wavFile.Bytes()
	if err != nil {
		t.Error(err)
	}
	os.WriteFile("../samples/easy_3.wav", fileBytes, 0644)
}
