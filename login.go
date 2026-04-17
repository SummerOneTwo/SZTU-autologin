package main

import (
	"encoding/json"
	"fmt"
	"io"
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

var (
	HTTPClient  = &http.Client{Timeout: 10 * time.Second}
	reChallenge = regexp.MustCompile(`"challenge":"(.*?)"`)
	reClientIP  = regexp.MustCompile(`"client_ip":"(.*?)"`)
	reErrorOK   = regexp.MustCompile(`"error":"ok"`)
	reJSONP     = regexp.MustCompile(`\((.*?)\)`)
	reQuote     = regexp.MustCompile(`'`)
)

// 创建带 User-Agent 的 GET 请求
func newGetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

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
	n      string
	typ    string
	enc    string
}

func NewLoginEngine(cfg Config) *LoginEngine {
	return &LoginEngine{
		config: cfg,
		acID:   cfg.ACID,
		n:      "200",
		typ:    "1",
		enc:    "srun_bx1",
	}
}

func (e *LoginEngine) getToken() error {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	params := url.Values{
		"callback": {fmt.Sprintf("jQuery112406608265734960486_%s", timestamp)},
		"username": {e.config.GetFullUsername()},
		"ip":       {e.ip},
		"_":        {timestamp},
	}

	requestURL := getChallengeAPI + "?" + params.Encode()

	req, err := newGetRequest(requestURL)
	if err != nil {
		return err
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}
	bodyStr := string(body)

	matches := reChallenge.FindStringSubmatch(bodyStr)
	if len(matches) < 2 {
		return fmt.Errorf("无法获取 challenge token")
	}
	e.token = matches[1]

	// 解析 client_ip（服务器看到的真实IP，避免本地VPN/代理导致的IP错误）
	ipMatches := reClientIP.FindStringSubmatch(bodyStr)
	if len(ipMatches) >= 2 {
		e.ip = ipMatches[1]
	}

	return nil
}

func (e *LoginEngine) getInfo() string {
	// 手动构建JSON以保持正确的键顺序：username, password, ip, acid, enc_ver
	return fmt.Sprintf(`{"username":"%s","password":"%s","ip":"%s","acid":"%s","enc_ver":"%s"}`,
		e.config.GetFullUsername(),
		e.config.Password,
		e.ip,
		e.acID,
		e.enc,
	)
}

// getChksum 生成校验字符串 - 与原始Python版本 get_chksum 完全一致
func (e *LoginEngine) getChksum() string {
	// 原始Python版本: token + username + token + hmd5 + token + ac_id + token + ip + token + n + token + type + token + i
	chkstr := e.token + e.config.GetFullUsername()
	chkstr += e.token + e.hmd5
	chkstr += e.token + e.acID
	chkstr += e.token + e.ip
	chkstr += e.token + e.n
	chkstr += e.token + e.typ
	chkstr += e.token + e.i
	return chkstr
}

func (e *LoginEngine) doComplexWork() error {
	infoStr := e.getInfo()
	e.i = "{SRBX1}" + getBase64(getXEncode(infoStr, e.token))
	e.hmd5 = getMD5(e.config.Password, e.token)
	chkstr := e.getChksum()
	e.chksum = getSHA1(chkstr)
	return nil
}

func (e *LoginEngine) Login() LoginResult {
	// 先获取 token 和 IP（从服务器响应中获取 client_ip，避免本地VPN/代理导致的IP错误）
	if err := e.getToken(); err != nil {
		return LoginResult{false, "获取token失败: " + err.Error()}
	}

	// 验证 IP 是否获取成功
	if e.ip == "" {
		return LoginResult{false, "无法获取本机IP"}
	}

	if err := e.doComplexWork(); err != nil {
		return LoginResult{false, "加密处理失败: " + err.Error()}
	}

	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// 手动构建 URL，确保 + 号不被编码
	// Python requests 库不会编码 + 号，但 Go 的 url.Values.Encode() 会
	// 注意：os 参数应该是 "windows 10"，requests 会将空格编码为 %20
	params := fmt.Sprintf("callback=jQuery11240645308969735664_%s&action=login&username=%s&password={MD5}%s&ac_id=%s&ip=%s&chksum=%s&info=%s&n=%s&type=%s&os=windows%%2010&name=windows&double_stack=0&_=%s",
		timestamp,
		url.QueryEscape(e.config.GetFullUsername()),
		e.hmd5,
		e.acID,
		e.ip,
		e.chksum,
		url.QueryEscape(e.i),
		e.n,
		e.typ,
		timestamp,
	)

	requestURL := srunPortalAPI + "?" + params

	req, err := newGetRequest(requestURL)
	if err != nil {
		return LoginResult{false, "创建请求失败: " + err.Error()}
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return LoginResult{false, "请求失败: " + err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LoginResult{false, "读取响应失败: " + err.Error()}
	}
	bodyStr := string(body)

	if reErrorOK.MatchString(bodyStr) {
		return LoginResult{true, "登录成功"}
	}

	matches := reJSONP.FindStringSubmatch(bodyStr)
	if len(matches) >= 2 {
		jsonStr := reQuote.ReplaceAllString(matches[1], `"`)
		var errData map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &errData); err == nil {
			errorCode, _ := errData["error"].(string)
			errorMsg, _ := errData["error_msg"].(string)
			if errorMsg != "" {
				return LoginResult{false, fmt.Sprintf("%s: %s", errorCode, errorMsg)}
			}
			return LoginResult{false, fmt.Sprintf("错误码: %s", errorCode)}
		}
	}

	return LoginResult{false, "登录失败: 无法解析响应"}
}
