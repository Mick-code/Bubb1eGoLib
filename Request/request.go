package Request

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type Request struct{
	Method string
	Url string
	Header map[string]string
	Body string
	PostValue map[string][]byte
	PostFile map[[2]string][]byte
	ProxyConf string
}


func InitRequest(Method string,Url string,Header map[string]string,Body string)* Request{
	r := Request{}
	r.Method = Method
	r.Url = Url
	r.Header = Header
	r.Body = Body
	return &r
}

func InitRequestByRaw(ip string,port string,security bool,requeststr string)* Request{
	r := Request{}
	headertmp := strings.Split(requeststr,"\n")
	first := strings.Split(headertmp[0]," ")
	method := first[0]
	uri := first[1]
	header := make(map[string]string)
	//获取header
	for k,v := range headertmp {
		if k==0 {
			continue
		}else if v==""{
			break;
		}
		tmp := strings.Split(v,": ")
		header[tmp[0]] = tmp[1]
	}
	//找body
	pos := strings.Index(requeststr,"\n\n")
	body := ""
	if(pos != -1 ){
		body = requeststr[pos:]
	}
	url := ""
	if security {
		url = "https://"
	}else{
		url = "http://"
	}
	url += ip + ":" + port + "/" + uri
	r.Method = method
	r.Url = url
	r.Header = header
	r.Body = body
	return &r
}

func (r * Request)Send()string{
	client := r.buildHttpClient()
	req,_ := http.NewRequest(r.Method,r.Url,strings.NewReader(r.Body))
	h := http.Header{}
	for k,v := range r.Header{
		h.Set(k,v)
	}
	req.Header = h
	resp,err := client.Do(req)
	if err !=nil {
		return err.Error()
	}
	defer resp.Body.Close()
	resbody, err := ioutil.ReadAll(resp.Body)
	if err !=nil {
		return err.Error()
	}
	return string(resbody)
}



func (r * Request)Post()string{
	client :=r.buildHttpClient()
	bodyBuf := &bytes.Buffer{}
	bodyWrite := multipart.NewWriter(bodyBuf)
	for k,v := range r.PostValue{
		fileWrite, err := bodyWrite.CreateFormField(k)
		if err!=nil{
			fmt.Println(err.Error())
		}
		fileWrite.Write(v)
	}


	for k,v := range r.PostFile{
		fileWrite, err := CreateFilePart(bodyWrite,k[0],k[1],"Content-Type: text/plain")
		if err!=nil{
			fmt.Println(err.Error())
		}
		tmpwriter := bufio.NewWriter(fileWrite)
		tmpwriter.Write(v)
		err = tmpwriter.Flush()
		if err != nil {
			fmt.Println(err)
		}
	}
	bodyWrite.Close()
	req,_ := http.NewRequest("POST",r.Url,bodyBuf)
	h := http.Header{}
	for k,v := range r.Header{
		h.Set(k,v)
	}
	contentType := bodyWrite.FormDataContentType()
	h.Set("Content-Type", contentType)
	req.Header = h

	resp,err := client.Do(req)
	if err !=nil {
		return err.Error()
	}
	defer resp.Body.Close()
	resbody, err := ioutil.ReadAll(resp.Body)
	if err !=nil {
		return err.Error()
	}
	return string(resbody)
}


func (r * Request)buildHttpClient() *http.Client {
	var proxy func(*http.Request) (*url.URL, error) = nil
	if r.ProxyConf!="" {
		proxy = func(_ *http.Request) (*url.URL, error) {
			return url.Parse("http://" + r.ProxyConf)
		}
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true},Proxy: proxy}
	client := &http.Client{Transport: transport}
	return client
}


func CreateFilePart(w *multipart.Writer,fieldname string, filename string,contentType string)(io.Writer, error){
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			quoteEscaper.Replace(fieldname), escapeQuotes(filename)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}