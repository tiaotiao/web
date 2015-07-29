package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

func ParseParams(c *Context) error {
	var err error
	var k, v string

	var r = c.Request

	// parse params in url path
	urlVars := mux.Vars(r)

	for k, v = range urlVars {
		c.Values[k] = v
	}

	contentType, contentParams, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		contentType = r.Header.Get("Content-Type")
		contentParams = nil
	}

	// parse params in url query
	// It can't parse post body because r.Body is empty now
	r.ParseForm()

	for k, _ = range r.Form {
		v = r.FormValue(k)
		c.Values[k] = v
	}

	// parse params in body
	if len(c.RawPostData) > 0 {

		if strings.Contains(contentType, "application/json") {
			var jsonValues = make(map[string]json.RawMessage) // just unpack the top layer of json struct

			err = json.Unmarshal(c.RawPostData, &jsonValues)

			if err != nil {
				return NewError(fmt.Sprintf("not json parameter: %s", err), http.StatusBadRequest)
			}

			for k, raw := range jsonValues {
				if len(raw) > 0 && (raw[0] != '"') {
					c.Values[k] = string(raw) // not unpack json slice or object
					continue
				}
				var v interface{}
				err = json.Unmarshal(raw, &v)
				if err != nil {
					return NewError(fmt.Sprintf("not json parameter: %s", err), http.StatusBadRequest)
				}
				c.Values[k] = v
			}

		} else if strings.Contains(contentType, "multipart/form-data") {
			buf := bytes.NewBuffer(c.RawPostData)

			mr := multipart.NewReader(buf, contentParams["boundary"])
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}

				d, err := ioutil.ReadAll(p)
				if err != nil {
					return err
				}

				part := new(struct {
					FormName string
					FileName string
					Header   textproto.MIMEHeader
					Data     []byte
				})
				part.FormName = p.FormName()
				part.FileName = p.FileName()
				part.Header = p.Header
				part.Data = d

				if part.FileName == "" {
					c.Values[part.FormName] = string(part.Data)
				}

				c.Multipart = append(c.Multipart, part)
			}

		} else if r.Method == "POST" ||
			r.Method == "DELETE" ||
			r.Method == "PUT" {

			var vals url.Values
			vals, err = url.ParseQuery(string(c.RawPostData))
			if err != nil {
				return fmt.Errorf("not querydict parameter")
			}

			for k, vs := range vals {
				if len(vs) == 0 {
					continue
				} else if len(vs) == 1 {
					c.Values[k] = vs[0]
				} else {
					c.Values[k] = strings.Join(vs, ",")
				}
			}
			c.Values["_POST_"] = vals
		}
	}

	return nil
}
