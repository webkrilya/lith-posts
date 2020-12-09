package lithLib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)
func getByMask(mask string, str string) string  {

	re := regexp.MustCompile(mask)
	res:= re.FindStringSubmatch(str)
	if (len(res)>0) {
		return res[0]
	} else {
		return ""
	}
}

func GetCookieForReg() ([]*http.Cookie, error) {
	uri := "https://r3.vfsglobal.com/LithuaniaAppt/Account/"
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Cookies(), nil
}

type regForm struct {
	DeText string
	Token string
	CapchaUri string
}

func GetRegForm(cook []*http.Cookie) ([]byte, error) {

	var answer regForm

	uri := "https://r3.vfsglobal.com/LithuaniaAppt/Account/RegisterUser?Length=7"
	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("upgrade-insecure-requests", "1")

	for _, v := range cook {
		req.AddCookie(v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	match := getByMask(`name="CaptchaDeText" type="hidden" value=".*?"`,string(body))
	answer.DeText = match[42:len(match)-1]

	match = getByMask(`__RequestVerificationToken" type="hidden" value=".*?"`,string(body))
	answer.Token =  match[49:len(match)-1]

	match = getByMask(`<img id="CaptchaImage" src=".*?"`,string(body))
	answer.CapchaUri = "https://r3.vfsglobal.com"+match[28:len(match)-1]

	js, e :=  json.Marshal(answer)
	if err != nil {
		return nil, e
	}

	return js, nil

}

func gUnzipData(data []byte) (resData []byte, err error) {
	b := bytes.NewBuffer(data)

	var r io.Reader
	r, err = gzip.NewReader(b)
	if err != nil {
		return
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return
	}
	resData = resB.Bytes()
	return
}

func GetCapcha(uri string, cook []*http.Cookie) (string,error) {

	req, err := http.NewRequest("GET", uri, nil)
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")
	req.Header.Set("accept","image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("accept-language","ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7");


	req.Header.Set("referer","https://r3.vfsglobal.com/LithuaniaAppt/Account/RegisterUser?Length=7")
	req.Header.Set("sec-fetch-dest","image")
	req.Header.Set("sec-fetch-mode","no-cors")
	req.Header.Set("sec-fetch-site","same-origin")
	req.Header.Set("accept-encoding","gzip, deflate, br")

	for _, v := range cook {
		req.AddCookie(v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil{
		return "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	uncompressedData, err := gUnzipData(body)
	if err!=nil{
		return "", err
	}

	capcha := base64.StdEncoding.EncodeToString(uncompressedData)
	return capcha, nil
}

func DoRegistration(js []byte, cook []*http.Cookie, username string, lastname string, mail string, phone string, password string) (string, error) {

	var jsStruct regForm

	err := json.Unmarshal(js,&jsStruct)
	if err!=nil{
		return "", err
	}

	uri := "https://r3.vfsglobal.com/LithuaniaAppt/Account/RegisterUser"

	params := url.Values{}

	params.Set("__RequestVerificationToken",jsStruct.Token);
	params.Set("IsGoogleCaptchaEnabled","False");
	params.Set("reCaptchaURL","https://www.google.com/recaptcha/api/siteverify?secret={0}&response={1}");
	params.Set("reCaptchaPublicKey","6Ld-Kg8UAAAAAK6U2Ur94LX8-Agew_jk1pQ3meJ1");
	params.Set("reCaptchaPrivateKey","6Ld-Kg8UAAAAAMX55MgXRFvSSeeeaI-aL9V8jvch");
	params.Set("FirstName",username);
	params.Set("LastName",lastname);
	params.Set("UserName",mail);
	params.Set("ContactNo",phone);
	params.Set("Password",password);
	params.Set("ConfirmPassword",password);
	params.Set("IsChecked","true");
	//params.Set("IsChecked","false");
	params.Set("CaptchaDeText",jsStruct.DeText);
	params.Set("CaptchaInputText", strings.ToUpper(jsStruct.CapchaUri));

	fmt.Println(params.Encode())

	req, err := http.NewRequest("POST", uri, strings.NewReader(params.Encode()))
	req.Header.Set("accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("accept-encoding","gzip, deflate, br")
	req.Header.Set("accept-language","ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("cache-control","max-age=0")
	req.Header.Set("content-length",strconv.Itoa(len(params.Encode())))
	req.Header.Set("content-type","application/x-www-form-urlencoded")
	req.Header.Set("origin","https://r3.vfsglobal.com")
	req.Header.Set("referer","https://r3.vfsglobal.com/LithuaniaAppt/Account/RegisterUser?Length=7")
	req.Header.Set("sec-fetch-dest","document")
	req.Header.Set("sec-fetch-mode","navigate")
	req.Header.Set("sec-fetch-site","same-origin")
	req.Header.Set("sec-fetch-user","?1")
	req.Header.Set("upgrade-insecure-requests","1")
	req.Header.Set("user-agent","Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")

	for _, v := range cook {
		req.AddCookie(v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err!=nil{
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err!=nil{
		return "", err
	}
	uncompressedData, err := gUnzipData(body)
	return string(uncompressedData), nil

}


