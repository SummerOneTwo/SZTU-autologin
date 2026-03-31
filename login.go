package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const (
	getChallengeAPI = "http://172.19.0.5/cgi-bin/get_challenge"
	srunPortalAPI   = "http://172.19.0.5/cgi-bin/srun_portal"
	userAgent       = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.26 Safari/537.36"
)

type LoginResult struct {
	Success bool
	Message string
}

type LoginEngine struct {
	config Config
	token  string
	ip     string
	acID   string
	i      string
	hmd5   string
	chksum string
}

func NewLoginEngine(cfg Config) *LoginEngine {
	return &LoginEngine{
		config: cfg,
		acID:   cfg.ACID,
	}
}

func (e *LoginEngine) getLocalIP() error {
	conn, err := net.Dial("udp", "172.19.0.5:80")
	if err != nil {
		return err
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	e.ip = localAddr.IP.String()
	return nil
}

func (e *LoginEngine) getToken() error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{
		"callback": {fmt.Sprintf("jQuery112406608265734960486_%s", timestamp)},
		"username": {e.config.GetFullUsername()},
		"ip":       {e.ip},
		"_":        {timestamp},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(getChallengeAPI + "?" + params.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`"challenge":"(.*?)"`)
	matches := re.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return fmt.Errorf("无法获取 challenge token")
	}
	e.token = matches[1]
	return nil
}

func (e *LoginEngine) doComplexWork() {
	info := map[string]string{
		"username": e.config.GetFullUsername(),
		"password": e.config.Password,
		"ip":       e.ip,
		"acid":     e.acID,
		"enc_ver":  "srun_bx1",
	}
	infoJSON, _ := json.Marshal(info)
	e.i = "{SRBX1}" + getBase64(getXEncode(string(infoJSON), e.token))
	e.hmd5 = getMD5(e.config.Password, e.token)

	chkstr := e.token + e.config.GetFullUsername() + e.token + e.hmd5 +
		e.token + e.acID + e.token + e.ip + e.token + "200" + e.token + "1" + e.token + e.i
	e.chksum = getSHA1(chkstr)
}

func (e *LoginEngine) Login() LoginResult {
	if err := e.getLocalIP(); err != nil {
		return LoginResult{false, "无法获取本机IP: " + err.Error()}
	}

	if err := e.getToken(); err != nil {
		return LoginResult{false, "获取token失败: " + err.Error()}
	}

	e.doComplexWork()

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{
		"callback":     {fmt.Sprintf("jQuery11240645308969735664_%s", timestamp)},
		"action":       {"login"},
		"username":     {e.config.GetFullUsername()},
		"password":     {"{MD5}" + e.hmd5},
		"ac_id":        {e.acID},
		"ip":           {e.ip},
		"chksum":       {e.chksum},
		"info":         {e.i},
		"n":            {"200"},
		"type":         {"1"},
		"os":           {"windows 10"},
		"name":         {"windows"},
		"double_stack": {"0"},
		"_":            {timestamp},
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(srunPortalAPI + "?" + params.Encode())
	if err != nil {
		return LoginResult{false, "请求失败: " + err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if regexp.MustCompile(`"error":"ok"`).MatchString(bodyStr) {
		return LoginResult{true, "登录成功"}
	}

	re := regexp.MustCompile(`\((.*?)\)`)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) >= 2 {
		jsonStr := regexp.MustCompile(`'`).ReplaceAllString(matches[1], `"`)
		var errData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &errData); err == nil {
			return LoginResult{false, fmt.Sprintf("%v: %v", errData["error"], errData["error_msg"])}
		}
	}

	return LoginResult{false, "登录失败: 无法解析响应"}
}
