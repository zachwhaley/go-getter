package getter

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

func TestHttpGetter_impl(t *testing.T) {
	var _ Getter = new(HttpGetter)
}

func TestHttpGetter_header(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/header"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_requestHeader(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	g.Header = make(http.Header)
	g.Header.Add("X-Foobar", "foobar")
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/expect-header"
	u.RawQuery = "expected=X-Foobar"

	// Get it!
	if err := g.GetFile(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "Hello\n")
}

func TestHttpGetter_meta(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_metaSubdir(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-subdir"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_metaSubdirGlob(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-subdir-glob"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_none(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/none"

	// Get it!
	if err := g.Get(dst, &u); err == nil {
		t.Fatal("should error")
	}
}

func TestHttpGetter_file(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempFile(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/file"

	// Get it!
	if err := g.GetFile(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "Hello\n")
}

func TestHttpGetter_ranged_request(t *testing.T) {
	var rangeReqs = []struct {
		Input    string
		Expected string
	}{
		{"0-", "Hello\n"},
		{"0-2", "Hel"},
		{"2-3", "ll"},
		{"3-", "lo\n"},
	}

	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempFile(t)
	for _, rangeReq := range rangeReqs {
		u, _ := urlhelper.Parse(fmt.Sprintf("http://%s/file?ranged_request_bytes=%s",
			ln.Addr().String(), rangeReq.Input))
		// Get it!
		if err := g.GetFile(dst, u); err != nil {
			t.Fatalf("err: %s", err)
		}

		// Verify the main file exists
		if _, err := os.Stat(dst); err != nil {
			t.Fatalf("err: %s", err)
		}
		assertContents(t, dst, rangeReq.Expected)
	}
}

func TestHttpGetter_auth(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta-auth"
	u.User = url.UserPassword("foo", "bar")

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_authNetrc(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/meta"

	// Write the netrc file
	path, closer := tempFileContents(t, fmt.Sprintf(testHttpNetrc, ln.Addr().String()))
	defer closer()
	defer tempEnv(t, "NETRC", path)()

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestHttpGetter_GetByteRange(t *testing.T) {
	var cases = []struct {
		Input     string // example: "5555-66666"
		Expected  []string
		ExpectErr bool
	}{
		// both start and end bytes given
		{"http://my/file.iso?ranged_request_bytes=555-666", []string{"555", "666"}, false},
		// no end given
		{"http://my/file.iso?ranged_request_bytes=555-", []string{"555", ""}, false},
		// end greater than start
		{"http://my/file.iso?ranged_request_bytes=666-555", nil, true},
		// unparseable into ints
		{"http://my/file.iso?ranged_request_bytes=66a6-5*55", nil, true},
		// no ranged request made
		{"http://my/file.iso", nil, false},
	}

	for _, testCase := range cases {
		src := testCase.Input
		u, err := urlhelper.Parse(src)
		byteRange, err := getByteRange(u)

		if byteRange == nil {
			if testCase.Expected != nil {
				t.Fatalf("unexpected output from getByteRange: expected %s; got %s",
					testCase.Expected, byteRange)
			}
		} else if byteRange[0] != testCase.Expected[0] && byteRange[1] != testCase.Expected[1] {
			t.Fatalf("unexpected output from getByteRange: expected %s; got %s",
				testCase.Expected, byteRange)
		}
		if err == nil && testCase.ExpectErr {
			t.Fatalf("Expected an error but didn't get one: %v", testCase)
		}
		if err != nil && !testCase.ExpectErr {
			t.Fatalf("Got an error we were not expecting: %v, %s", testCase,
				err)
		}
	}
}

// test round tripper that only returns an error
type errRoundTripper struct{}

func (errRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return nil, errors.New("test round tripper")
}

// verify that the default httpClient no longer comes from http.DefaultClient
func TestHttpGetter_cleanhttp(t *testing.T) {
	ln := testHttpServer(t)
	defer ln.Close()

	// break the default http client
	http.DefaultClient.Transport = errRoundTripper{}
	defer func() {
		http.DefaultClient.Transport = http.DefaultTransport
	}()

	g := new(HttpGetter)
	dst := tempDir(t)
	defer os.RemoveAll(dst)

	var u url.URL
	u.Scheme = "http"
	u.Host = ln.Addr().String()
	u.Path = "/header"

	// Get it!
	if err := g.Get(dst, &u); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testHttpServer(t *testing.T) net.Listener {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/expect-header", testHttpHandlerExpectHeader)
	mux.HandleFunc("/file", testHttpHandlerFile)
	mux.HandleFunc("/header", testHttpHandlerHeader)
	mux.HandleFunc("/meta", testHttpHandlerMeta)
	mux.HandleFunc("/meta-auth", testHttpHandlerMetaAuth)
	mux.HandleFunc("/meta-subdir", testHttpHandlerMetaSubdir)
	mux.HandleFunc("/meta-subdir-glob", testHttpHandlerMetaSubdirGlob)

	var server http.Server
	server.Handler = mux
	go server.Serve(ln)

	return ln
}

func testHttpHandlerExpectHeader(w http.ResponseWriter, r *http.Request) {
	if expected, ok := r.URL.Query()["expected"]; ok {
		if r.Header.Get(expected[0]) != "" {
			w.Write([]byte("Hello\n"))
			return
		}
	}

	w.WriteHeader(400)
}

func testHttpHandlerFile(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "hello.txt", time.Now(), strings.NewReader("Hello\n"))
}

func testHttpHandlerHeader(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("X-Terraform-Get", testModuleURL("basic").String())
	w.WriteHeader(200)
}

func testHttpHandlerMeta(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaAuth(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(401)
		return
	}

	if user != "foo" || pass != "bar" {
		w.WriteHeader(401)
		return
	}

	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic").String())))
}

func testHttpHandlerMetaSubdir(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic//subdir").String())))
}

func testHttpHandlerMetaSubdirGlob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf(testHttpMetaStr, testModuleURL("basic//sub*").String())))
}

func testHttpHandlerNone(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(testHttpNoneStr))
}

const testHttpMetaStr = `
<html>
<head>
<meta name="terraform-get" content="%s">
</head>
</html>
`

const testHttpNoneStr = `
<html>
<head>
</head>
</html>
`

const testHttpNetrc = `
machine %s
login foo
password bar
`
