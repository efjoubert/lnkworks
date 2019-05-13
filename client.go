package lnksworks

import (
	"io"
	"strings"

	"github.com/efjoubert/lnkworks/httpclient"
)

//Client conveniance struct wrapping arround *http.Client, io.Writer -> io.Reader ...
type Client struct {
	httpclnt  *httpclient.HttpClient
	params    *Parameters
	reqhders  *RequestHeader
	reqcntnt  *RequestContent
	resphders *ResponseHeader
	respcntnt *ResponseContent
	atvpros   *ActiveProcessor
}

//Header core Header map
type Header map[string][]string

//Value Header
func (hdr Header) Value(name string) []string {
	return hdr[name]
}

//Append values(s)
func (hdr Header) Append(name string, value ...string) {
	if len(value) > 0 {
		var vals = hdr[name]
		if vals == nil {
			vals = []string{}
		}
		vals = append(vals, value...)
		hdr[name] = vals[:]
		vals = nil
	}
}

//Keys []string of keys
func (hdr Header) Keys() (keys []string) {
	for k := range hdr {
		keys = append(keys, k)
	}
	return
}

//ContainsKey check if Header ContainsKey
func (hdr Header) ContainsKey(key string) (keyok bool) {
	_, keyok = hdr[key]
	return
}

//SetValue Header
func (hdr Header) SetValue(name string, value ...string) {
	hdr[name] = value[:]
}

//Values [][]string values
func (hdr Header) Values(name ...string) (values [][]string) {
	for nme, val := range hdr {
		if len(name) > 0 {
			for _, nm := range name {
				if strings.ToUpper(nm) == strings.ToUpper(nme) {
					values = append(values, val[:])
				}
			}
		}
	}

	return
}

//RequestHeader request Header
type RequestHeader struct {
	*Header
	clnt *Client
}

//ReadAll  method to populate (write) RequestHeader from io.Reader
func (reqHdr *RequestHeader) ReadAll(r io.Reader, keyvalsep string, headersep string) (err error) {
	if keyvalsep == "" {
		keyvalsep = ":"
	}
	if headersep == "" {
		headersep = "r\n"
	}
	var headersepbytes = []byte(headersep[:])
	var headersepi = 0
	var p = make([]byte, 4096)
	var header = ""
	var val = ""
	var vals = []string{}
	var valsi = 0
	var valsL = 0

	//0 = header
	//1 = value
	var readStage = 0
	for {
		if n, nerr := r.Read(p[:]); nerr == nil {
			if n == 0 {
				err = io.EOF
				break
			} else if n > 0 {
				for pi := range p[:n] {
					if readStage == 0 {
						if string(p[pi:pi]) != "" {
							if headersepi < len(headersepbytes) && p[pi] == headersepbytes[headersepi] {
								headersepi++
								if headersepi == len(headersepbytes) {
									break
								}
							} else {
								if string(p[pi:pi]) == keyvalsep {
									readStage++
								} else {
									header += string(p[pi:pi])
								}
							}
						}
					} else if readStage == 1 {
						if headersepi < len(headersepbytes) && p[pi] == headersepbytes[headersepi] {
							headersepi++
							if headersepi == len(headersepbytes) {
								valsi = 0
								for _, v := range strings.Split(strings.TrimSpace(val), ";") {
									if v = strings.TrimSpace(v); v != "" {
										if valsi < valsL {
											vals[valsi] = v
										} else {
											vals = append(vals, v)
											valsL++
										}
										valsi++
									}
								}
								if valsL < len(vals) {
									valsL = len(vals)
								}
								if valsi > 0 {
									reqHdr.Append(header, vals[:valsi]...)
								}
								header = ""
								val = ""
								headersepi = 0
							}
						} else if headersepi == 0 {
							val += string(p[pi:pi])
						}
					}
				}
			}
		} else if nerr != nil {
			if nerr != io.EOF {
				err = nil
			}
			break
		}
	}
	return
}

//Client public Client method for RequestHeader
func (reqHdr *RequestHeader) Client() *Client {
	return reqHdr.clnt
}

//RequestContent Client RequestContent
type RequestContent struct {
	clnt *Client
}

//ResponseHeader response Header
type ResponseHeader struct {
	*Header
	clnt *Client
}

//ReadAll  method to populate (write) ResponseHeader from io.Reader
func (respHdr *ResponseHeader) ReadAll(r io.Reader, keyvalsep string, headersep string) (err error) {
	if keyvalsep == "" {
		keyvalsep = ":"
	}
	if headersep == "" {
		headersep = "r\n"
	}
	var headersepbytes = []byte(headersep[:])
	var headersepi = 0
	var p = make([]byte, 4096)
	var header = ""
	var val = ""
	var vals = []string{}
	var valsi = 0
	var valsL = 0

	//0 = header
	//1 = value
	var readStage = 0
	for {
		if n, nerr := r.Read(p[:]); nerr == nil {
			if n == 0 {
				err = io.EOF
				break
			} else if n > 0 {
				for pi := range p[:n] {
					if readStage == 0 {
						if string(p[pi:pi]) != "" {
							if headersepi < len(headersepbytes) && p[pi] == headersepbytes[headersepi] {
								headersepi++
								if headersepi == len(headersepbytes) {
									break
								}
							} else {
								if string(p[pi:pi]) == keyvalsep {
									readStage++
								} else {
									header += string(p[pi:pi])
								}
							}
						}
					} else if readStage == 1 {
						if headersepi < len(headersepbytes) && p[pi] == headersepbytes[headersepi] {
							headersepi++
							if headersepi == len(headersepbytes) {
								valsi = 0
								for _, v := range strings.Split(strings.TrimSpace(val), ";") {
									if v = strings.TrimSpace(v); v != "" {
										if valsi < valsL {
											vals[valsi] = v
										} else {
											vals = append(vals, v)
											valsL++
										}
										valsi++
									}
								}
								if valsL < len(vals) {
									valsL = len(vals)
								}
								if valsi > 0 {
									respHdr.Append(header, vals[:valsi]...)
								}
								header = ""
								val = ""
								headersepi = 0
							}
						} else if headersepi == 0 {
							val += string(p[pi:pi])
						}
					}
				}
			}
		} else if nerr != nil {
			if nerr != io.EOF {
				err = nil
			}
			break
		}
	}
	return
}

//Client public Client method for ResponseHeader
func (respHdr *ResponseHeader) Client() *Client {
	return respHdr.clnt
}

//ResponseContent Client ResponseContent
type ResponseContent struct {
	clnt *Client
}

//NewClient instantiate
func NewClient(r io.Reader, w io.Writer) (clnt *Client) {
	clnt = &Client{atvpros: NewActiveProcessor(w)}
	clnt.params = NewParameters()
	clnt.reqhders = &RequestHeader{clnt: clnt}
	clnt.reqcntnt = &RequestContent{clnt: clnt}
	clnt.resphders = &ResponseHeader{clnt: clnt}
	clnt.respcntnt = &ResponseContent{clnt: clnt}
	return
}

//CleanupClient cleanup Client instance
func (clnt *Client) CleanupClient() {
	if clnt.atvpros != nil {
		clnt.atvpros.cleanupActiveProcessor()
		clnt.atvpros = nil
	}
	if clnt.httpclnt != nil {
		clnt.httpclnt = nil
	}
	if clnt.params != nil {
		clnt.params.CleanupParameters()
		clnt.params = nil
	}
	if clnt.reqcntnt != nil {
		clnt.reqcntnt = nil
	}
	if clnt.reqhders != nil {
		clnt.reqhders = nil
	}
	if clnt.respcntnt != nil {
		clnt.respcntnt = nil
	}
	if clnt.resphders != nil {
		clnt.resphders = nil
	}
}
