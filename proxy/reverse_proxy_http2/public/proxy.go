package public

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/e421083458/gateway_demo/proxy/reverse_proxy_http2/testdata"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var transport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, //连接超时
		KeepAlive: 30 * time.Second, //长连接超时时间
	}).DialContext,
	//TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	TLSClientConfig: func() *tls.Config {
		pool := x509.NewCertPool()
		caCertPath := testdata.Path("ca.crt")
		caCrt, _ := ioutil.ReadFile(caCertPath)
		pool.AppendCertsFromPEM(caCrt)
		return &tls.Config{RootCAs: pool}
	}(),
	MaxIdleConns:          100,              //最大空闲连接
	IdleConnTimeout:       90 * time.Second, //空闲超时时间
	TLSHandshakeTimeout:   10 * time.Second, //tls握手超时时间
	ExpectContinueTimeout: 1 * time.Second,  //100-continue 超时时间
}

func NewMultipleHostsReverseProxy(targets []*url.URL) *httputil.ReverseProxy {
	//请求协调者
	director := func(req *http.Request) {
		targetIndex := rand.Intn(len(targets))
		target := targets[targetIndex]
		targetQuery := target.RawQuery
		fmt.Println("target.Scheme")
		fmt.Println(target.Scheme)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "user-agent")
		}
	}
	return &httputil.ReverseProxy{Director: director, Transport: transport,}
	////更改内容
	//modifyFunc := func(resp *http.Response) error {
	//	//请求以下命令：curl 'http://127.0.0.1:2002/error'
	//	if resp.StatusCode != 200 {
	//		//获取内容
	//		oldPayload, err := ioutil.ReadAll(resp.Body)
	//		if err != nil {
	//			return err
	//		}
	//		//追加内容
	//		newPayload := []byte("StatusCode error:" + string(oldPayload))
	//		resp.Body = ioutil.NopCloser(bytes.NewBuffer(newPayload))
	//		resp.ContentLength = int64(len(newPayload))
	//		resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(newPayload)), 10))
	//	}
	//	return nil
	//}
	//
	////错误回调 ：关闭real_server时测试，错误回调
	////范围：transport.RoundTrip发生的错误、以及ModifyResponse发生的错误
	//errFunc := func(w http.ResponseWriter, r *http.Request, err error) {
	//	http.Error(w, "ErrorHandler error:"+err.Error(), 500)
	//}
	//
	//return &httputil.ReverseProxy{Director: director, Transport: transport, ModifyResponse: modifyFunc, ErrorHandler: errFunc}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
