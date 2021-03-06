/* A few of the below functions are heavilly influenced by OJ's code in gobuster */

package libgrequest

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	// "encoding/base64"
)

// RedirectHandler :
type RedirectHandler struct {
	Transport http.RoundTripper
	State     *State
}

// RedirectError : redirect err struct from gobuster
type RedirectError struct {
	StatusCode int
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("%-8d", e.StatusCode)
}

// RoundTrip : roundtrip from gobuster
func (rh *RedirectHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if rh.State.FollowRedirect {
		return rh.Transport.RoundTrip(req)
	}
	resp, err = rh.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusNotModified, http.StatusUseProxy, http.StatusTemporaryRedirect:
		return nil, &RedirectError{StatusCode: resp.StatusCode}
	}
	return resp, err
}

// TODO: less redundant code
func makePostFormRequest(s *State, fullURL, cookie, cmdline string) (*int, error) {
	// fmt.Println(payload)
	s.Counter.Inc()
	var patpostform = "--post-form [^\t\n\f\r ]+"
	postform := regexp.MustCompile(patpostform)
	payload := postform.FindString(cmdline)[len("--post-form "):]
	v := url.Values{}
	pairs := strings.Split(payload, ",")
	for i := range pairs {
		kv := strings.Split(pairs[i], "=")
		if len(kv) == 2 {
			v.Set(kv[0], kv[1])
		}
	}
	encv := v.Encode()
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(encv))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if cookie != "" {
		if s.Fuzzer.Cmdline[1] {
			cookie = ArgString(cmdline, "-b [^\t\n\f\r ]+")[len("-b "):]
		} else {
			req.Header.Set("Cookie", cookie)
		}
	}
	if s.UserAgent != "" {
		if s.Fuzzer.Cmdline[4] {
			s.UserAgent = ArgString(cmdline, "-ua [^\t\n\f\r ]+")[len("-ua."):]
		} else {
			req.Header.Set("User-Agent", s.UserAgent)
		}
	}
	if s.Username != "" {
		if s.Fuzzer.Cmdline[3] {
			s.Username = ArgString(cmdline, "--username [^\t\n\f\r ]+")[len("--username."):]
		}
		if s.Fuzzer.Cmdline[2] {
			s.Password = ArgString(cmdline, "--password [^\t\n\f\r ]+")[len("--password."):]
		}
		req.SetBasicAuth(s.Username, s.Password)
	}
	if len(s.Headers) > 0 {
		for k, v := range s.Headers {
			// fmt.Println("key", v, "value", v)
			req.Header.Set(k, v)
		}
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[!] Invalid certificate, try using -k.")
			}
			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}
	defer resp.Body.Close()

	r, err := InitResult(payload, resp)
	if err != nil {
		return nil, nil
	}
	// Print Output
	if s.Quiet != true {
		s.Printer(s, r)
	}
	return &resp.StatusCode, nil
}

func makePostMultiRequest(s *State, fullURL, cookie, cmdline string) (*int, error) {
	s.Counter.Inc()
	var patmultpart = "--post-multipart [^\t\n\f\r ]+"
	multpartform := regexp.MustCompile(patmultpart)
	var payload = multpartform.FindString(cmdline)[len("--post-multipart "):]
	var err error
	values := map[string]io.Reader{
		// "file":  mustOpen("main.go"), // lets assume its this file
		payload: strings.NewReader(""),
	}
	var b bytes.Buffer
	multipartw := multipart.NewWriter(&b)
	for key, rdr := range values {
		var fwtr io.Writer
		if f, ok := rdr.(io.Closer); ok {
			defer f.Close()
		}
		if f, ok := rdr.(*os.File); ok {
			if fwtr, _ = multipartw.CreateFormFile(key, f.Name()); err != nil {
				return nil, nil
			}
		} else {
			if fwtr, err = multipartw.CreateFormField(key); err != nil {
				return nil, nil
			}
		}
		if _, err = io.Copy(fwtr, rdr); err != nil {
			return nil, err
		}
	}
	multipartw.Close()
	req, err := http.NewRequest("POST", fullURL, &b)
	if err != nil {
		return nil, nil
	}
	req.Header.Add("Content-Type", multipartw.FormDataContentType())
	if cookie != "" {
		if s.Fuzzer.Cmdline[1] {
			cookie = ArgString(cmdline, "-b [^\t\n\f\r ]+")[len("-b "):]
		} else {
			req.Header.Set("Cookie", cookie)
		}
	}
	if s.UserAgent != "" {
		if s.Fuzzer.Cmdline[4] {
			s.UserAgent = ArgString(cmdline, "-ua [^\t\n\f\r ]+")[len("-ua."):]
		} else {
			req.Header.Set("User-Agent", s.UserAgent)
		}
	}
	if s.Username != "" {
		if s.Fuzzer.Cmdline[3] {
			s.Username = ArgString(cmdline, "--username [^\t\n\f\r ]+")[len("--username."):]
		}
		if s.Fuzzer.Cmdline[2] {
			s.Password = ArgString(cmdline, "--password [^\t\n\f\r ]+")[len("--password."):]
		}
		req.SetBasicAuth(s.Username, s.Password)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[!] Invalid certificate, try using -k.")
			}
			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}
	defer resp.Body.Close()

	r, err := InitResult(payload, resp)
	if err != nil {
		return nil, nil
	}
	// Print Output
	if s.Quiet != true {
		s.Printer(s, r)
	}
	return &resp.StatusCode, nil
}

// makeRequest : make http request
func makeRequest(s *State, fullURL, cookie, cmdline string) (*int, error) {
	s.Counter.Inc()
	//fmt.Println("URL:", fullURL, "cookies:", cookie,"payload:", cmdline)

	if s.Fuzzer.Cmdline[0] {
		fullURL = ArgString(cmdline, "htt(p|ps)[^\t\n\f\r ]+$")
	}
	req, err := http.NewRequest(s.Method, fullURL, nil)
	if err != nil {
		return nil, nil
	}
	if cookie != "" {
		if s.Fuzzer.Cmdline[1] {
			cookie = ArgString(cmdline, "-b [^\t\n\f\r ]+")[len("-b "):]
		} else {
			req.Header.Set("Cookie", cookie)
		}
	}
	if s.UserAgent != "" {
		if s.Fuzzer.Cmdline[4] {
			s.UserAgent = ArgString(cmdline, "-ua [^\t\n\f\r ]+")[len("-ua."):]
		} else {
			req.Header.Set("User-Agent", s.UserAgent)
		}
	}
	if s.Username != "" {
		if s.Fuzzer.Cmdline[3] {
			s.Username = ArgString(cmdline, "--username [^\t\n\f\r ]+")[len("--username."):]
		}
		if s.Fuzzer.Cmdline[2] {
			s.Password = ArgString(cmdline, "--password [^\t\n\f\r ]+")[len("--password."):]
		}
		req.SetBasicAuth(s.Username, s.Password)
	}
	if len(s.Headers) > 0 {
		for k, v := range s.Headers {
			fmt.Println("key", v, "value", v)
			req.Header.Set(k, v)
		}
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[!] Invalid certificate, try using -k.")
			}
			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}
	defer resp.Body.Close()

	r, err := InitResult(fullURL, resp)
	if err != nil {
		return nil, nil
	}
	// Print Output
	if s.Quiet != true {
		s.Printer(s, r)
	}
	return &resp.StatusCode, nil
}

// GoGet : returs address of response statuscode and error
func GoGet(s *State, url, cookie, payload string) (*int, error) {
	return makeRequest(s, url, cookie, payload)
}

// GoPostForm :
func GoPostForm(s *State, url, cookie, payload string) (*int, error) {
	return makePostFormRequest(s, url, cookie, payload)
}

// GoPostMultiPart :
func GoPostMultiPart(s *State, url, cookie, payload string) (*int, error) {
	return makePostMultiRequest(s, url, cookie, payload)
}
