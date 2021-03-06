package rain

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

var data = []string{"asdf", "a", "", "qwerty"}

func TestFiles(t *testing.T) {
	dir, err := ioutil.TempDir("", "partialfile-")
	if err != nil {
		t.Fatal(err)
	}
	for i, s := range data {
		filename := filepath.Join(dir, "file"+strconv.Itoa(i))
		err = ioutil.WriteFile(filename, []byte(s), 0666)
		if err != nil {
			t.Fatal(err)
		}
	}
	osFiles := make([]*os.File, len(data))
	for i := range data {
		filename := filepath.Join(dir, "file"+strconv.Itoa(i))
		osFiles[i], err = os.OpenFile(filename, os.O_RDWR, 0666)
		if err != nil {
			t.Fatal(err)
		}
	}
	files := []section{
		section{osFiles[0], 2, 2},
		section{osFiles[1], 0, 1},
		section{osFiles[2], 0, 0},
		section{osFiles[3], 0, 2},
	}
	pf := sections(files)
	b := make([]byte, 5)
	n, err := io.ReadFull(pf.Reader(), b)
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Errorf("n == %d", n)
	}
	if string(b) != "dfaqw" {
		t.Errorf("b = %s", string(b))
	}
	n, err = pf.Write([]byte("12345"))
	if err != nil {
		t.Error(err)
	}
	if n != 5 {
		t.Errorf("n == %d", n)
	}
	if content(osFiles[0]) != "as12" {
		t.Fail()
	}
	if content(osFiles[1]) != "3" {
		t.Fail()
	}
	if content(osFiles[2]) != "" {
		t.Fail()
	}
	if content(osFiles[3]) != "45erty" {
		t.Fail()
	}
}

func content(f *os.File) string {
	f.Seek(0, os.SEEK_SET)
	fi, _ := f.Stat()
	b := make([]byte, fi.Size())
	f.Read(b)
	return string(b)
}
