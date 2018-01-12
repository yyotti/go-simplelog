package slog

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"testing"
)

func tempFile() *os.File {
	f, err := ioutil.TempFile(os.TempDir(), "new")
	if err != nil {
		panic(err)
	}

	return f
}

func captureStd(f func()) (string, string) {
	stdout := os.Stdout
	stderr := os.Stderr

	ro, wo, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = wo

	re, we, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = we

	f()

	wo.Close()
	we.Close()
	os.Stdout = stdout
	os.Stderr = stderr

	outo, err := ioutil.ReadAll(ro)
	if err != nil {
		panic(err)
	}
	ro.Close()

	oute, err := ioutil.ReadAll(re)
	if err != nil {
		panic(err)
	}
	re.Close()

	return string(outo), string(oute)
}

func TestInitDefault(t *testing.T) {
	l = nil

	stdout, stderr := captureStd(func() {
		initDefault()

		if l == nil {
			t.Fatal("Default logger is nil")
		}

		l.infoLogger.Print("Info")
		l.errorLogger.Print("Error")
	})

	infoReg := regexp.MustCompile(`\[info\]  \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} (.*)\n`)
	m := infoReg.FindStringSubmatch(stdout)
	if len(m) == 0 {
		t.Errorf("Invalid log format:%s", stdout)
	} else if m[1] != "Info" {
		t.Errorf("Invalid info message:%s", stdout)
	}

	errorReg := regexp.MustCompile(`\[error\] \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} (.*)\n`)
	m = errorReg.FindStringSubmatch(stderr)
	if len(m) == 0 {
		t.Errorf("Invalid log format:%s", stderr)
	} else if m[1] != "Error" {
		t.Errorf("Invalid error message:%s", stderr)
	}
}

// FIXME cyclomatic complexity is high
// nolint:gocyclo
func TestInit(t *testing.T) {
	tests := []struct {
		debug bool
		flag  int
		files []io.Writer
	}{
		// TEST0 {{{
		{
			debug: false,
			files: nil,
		},
		// }}}
		// TEST1 {{{
		{
			debug: true,
			files: nil,
		},
		// }}}
		// TEST2 {{{
		{
			debug: false,
			files: []io.Writer{tempFile()},
		},
		// }}}
	}

	for i, tt := range tests {
		l = nil

		for _, f := range tt.files {
			defer func(f *os.File) {
				f.Close()
				os.Remove(f.Name())
			}(f.(*os.File))
		}

		debugMsg := ""
		infoMsg := "Info test"
		errorMsg := "Error test"

		stdout, stderr := captureStd(func() {
			Init(tt.debug, tt.flag, tt.files...)

			if l.infoLogger == nil {
				t.Fatalf("%d: infoLogger is nil", i)
			}
			if l.errorLogger == nil {
				t.Fatalf("%d: errorLogger is nil", i)
			}
			if tt.debug {
				if l.debugLogger == nil {
					t.Fatalf("%d: debugLogger is nil", i)
				}
			} else {
				if l.debugLogger != nil {
					t.Fatalf("%d: debugLogger is not nil", i)
				}
			}

			l.infoLogger.SetPrefix("")
			l.infoLogger.SetFlags(0)
			l.infoLogger.Println(infoMsg)

			l.errorLogger.SetPrefix("")
			l.errorLogger.SetFlags(0)
			l.errorLogger.Println(errorMsg)

			if tt.debug {
				debugMsg = "Debug test"

				l.debugLogger.SetPrefix("")
				l.debugLogger.SetFlags(0)
				l.debugLogger.Println(debugMsg)
			}
		})

		expo := infoMsg + "\n"
		if tt.debug {
			expo += debugMsg + "\n"
		}
		if string(stdout) != expo {
			t.Errorf("%d: Expected [%s] for STDOUT, but got [%s]", i, expo, string(stdout))
		}

		expe := errorMsg + "\n"
		if string(stderr) != expe {
			t.Errorf("%d: Expected [%s] for STDERR, but got [%s]", i, expe, string(stderr))
		}

		expf := infoMsg + "\n" + errorMsg + "\n"
		for _, f := range tt.files {
			file := f.(*os.File)
			err := file.Close()
			if err != nil {
				t.Errorf("Err!%v", err)
			}

			outf, err := ioutil.ReadFile(file.Name())
			if err != nil {
				panic(err)
			}

			if string(outf) != expf {
				t.Errorf("%d: Expected [%s] for file %s, but got [%s]", i, expf, f.(*os.File).Name(), string(outf))
			}
		}
	}
}

func TestWrite(t *testing.T) {
	tests := []struct {
		v        []interface{}
		expected string
	}{
		// TEST0 {{{
		{
			v:        nil,
			expected: "",
		},
		// }}}
		// TEST1 {{{
		{
			v:        []interface{}{"a\n"},
			expected: "a",
		},
		// }}}
		// TEST2 {{{
		{
			v:        []interface{}{"a", "b"},
			expected: "ab",
		},
		// }}}
		// TEST3 {{{
		{
			v:        []interface{}{"%02d\n", 1},
			expected: "%02d\n1",
		},
		// }}}
	}

	var buf bytes.Buffer
	for i, tt := range tests {
		buf.Reset()
		l := log.New(&buf, "", 0)
		write(l, tt.v...)

		expected := tt.expected + "\n"
		actual := buf.String()
		if actual != expected {
			t.Errorf("%d: Expected get [%s], but got [%s]", i, expected, actual)
		}
	}
}

func TestWritef(t *testing.T) {
	tests := []struct {
		format   string
		v        []interface{}
		expected string
	}{
		// TEST0 {{{
		{
			format:   "",
			v:        nil,
			expected: "",
		},
		// }}}
		// TEST1 {{{
		{
			format:   "a\n",
			v:        nil,
			expected: "a",
		},
		// }}}
		// TEST2 {{{
		{
			format:   "%sabc",
			v:        []interface{}{"%02d\n"},
			expected: "%02d\nabc",
		},
		// }}}
		// TEST3 {{{
		{
			format:   "%02d\n%s",
			v:        []interface{}{1, "abc"},
			expected: "01\nabc",
		},
		// }}}
	}

	var buf bytes.Buffer
	for i, tt := range tests {
		buf.Reset()
		l := log.New(&buf, "", 0)
		writef(l, tt.format, tt.v...)

		expected := tt.expected + "\n"
		actual := buf.String()
		if actual != expected {
			t.Errorf("%d: Expected get [%s], but got [%s]", i, expected, actual)
		}
	}
}

func TestWriteln(t *testing.T) {
	tests := []struct {
		v        []interface{}
		expected string
	}{
		// TEST0 {{{
		{
			v:        nil,
			expected: "",
		},
		// }}}
		// TEST1 {{{
		{
			v:        []interface{}{"a"},
			expected: "a",
		},
		// }}}
		// TEST2 {{{
		{
			v:        []interface{}{"%02d\n", "abc"},
			expected: "%02d\n abc",
		},
		// }}}
		// TEST3 {{{
		{
			v:        []interface{}{1, "abc\n"},
			expected: "1 abc\n",
		},
		// }}}
	}

	var buf bytes.Buffer
	for i, tt := range tests {
		buf.Reset()
		l := log.New(&buf, "", 0)
		writeln(l, tt.v...)

		expected := tt.expected + "\n"
		actual := buf.String()
		if actual != expected {
			t.Errorf("%d: Expected get [%s], but got [%s]", i, expected, actual)
		}
	}
}

// FIXME Add tests
