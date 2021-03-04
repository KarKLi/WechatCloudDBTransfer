package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

const (
	// port Server端接收请求的本地端口
	port = "20000"
	// OldAppSecret 旧版小程序的AppSecret
	OldAppSecret = ""
	// OldAppID 旧版小程序的AppID
	OldAppID = ""
	// NewAppSecret 新版小程序的AppSecret
	NewAppSecret = ""
	// NewAppID 新版小程序的AppID
	NewAppID = ""
	// Oldenv 旧版小程序的云数据库环境ID
	Oldenv = ""
	// Oldenv 新版小程序的云数据库环境ID
	Newenv = ""
)

type accessTokenRsp struct {
	AccessToken string `json:"access_token"`
	ExpireIn    int32  `json:"expire_in"`
	ErrCode     int32  `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type cloudFileURLRsp struct {
	ErrCode  int32  `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
	FileList []struct {
		FileID      string `json:"fileid"`
		DownloadURL string `json:"download_url"`
		Status      int32  `json:"status"`
	} `json:"file_list"`
}

type cloudFileUploadFirstRsp struct {
	ErrCode       int32  `json:"errcode"`
	ErrMsg        string `json:"errmsg"`
	URL           string `json:"url"`
	Token         string `json:"token"`
	Authorization string `json:"authorization"`
	FileID        string `json:"file_id"`
	CosFileID     string `json:"cos_file_id"`
}

type cloudFileUploadSecondRsp struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func GetAccessToken(AppID string, AppSecret string) string {
	resp, err := http.Get(fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", AppID, AppSecret))
	if err != nil {
		panic("http.Get access_token failed.")
	}
	rspBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Read access_token from response failed.")
	}
	rsp := &accessTokenRsp{}
	err = json.Unmarshal(rspBytes, rsp)
	if err != nil {
		panic("Unmarshal access_token rsp failed.")
	}
	if rsp.ErrCode != 0 {
		panic(fmt.Sprintf("Remote call failed, errcode:%d, errmsg:%s", rsp.ErrCode, rsp.ErrMsg))
	}
	return rsp.AccessToken
}
func GetDownloadURL(AccessToken string, FileID string) (URL string, err error) {
	resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/tcb/batchdownloadfile?access_token=%s", AccessToken), "application/json", strings.NewReader(
		fmt.Sprintf(
			`
		{
			"env": "%s",
			"file_list": [
				{
					"fileid": "%s",
					"max_age": 7200
				}
			]
		}
		`, Oldenv, FileID)))
	if err != nil {
		return "", errors.New("Weixin API not avaliable.")
	}
	rspBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Read download URL from response failed.")
	}
	rsp := &cloudFileURLRsp{}
	err = json.Unmarshal(rspBytes, rsp)
	if err != nil {
		return "", errors.New("Unmarshal access_token rsp failed.")
	}
	if rsp.ErrCode != 0 {
		return "", errors.New(fmt.Sprintf("Remote call failed, errcode:%d, errmsg:%s", rsp.ErrCode, rsp.ErrMsg))
	}
	return rsp.FileList[0].DownloadURL, nil
}
func UploadFile(AccessToken string, Path string, DownloadURL string) error {
	// 先执行第一次POST
	resp, err := http.Post(fmt.Sprintf("https://api.weixin.qq.com/tcb/uploadfile?access_token=%s", AccessToken), "application/json", strings.NewReader(
		fmt.Sprintf(
			`
		{
			"env": "%s",
			"path": "%s"
		}
		`, Newenv, Path)))
	if err != nil {
		return err
	}
	rspBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Read response failed.")
	}
	rsp := &cloudFileUploadFirstRsp{}
	err = json.Unmarshal(rspBytes, rsp)
	if err != nil {
		return errors.New("Unmarshal access_token rsp failed.")
	}
	if rsp.ErrCode != 0 {
		return errors.New(fmt.Sprintf("Remote call failed, errcode:%d, errmsg:%s", rsp.ErrCode, rsp.ErrMsg))
	}
	// 拼接第二个POST
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	err = w.WriteField("key", Path)
	if err != nil {
		return err
	}
	err = w.WriteField("Signature", rsp.Authorization)
	if err != nil {
		return err
	}
	err = w.WriteField("x-cos-security-token", rsp.Token)
	if err != nil {
		return err
	}
	err = w.WriteField("x-cos-meta-fileid", rsp.CosFileID)
	if err != nil {
		return err
	}
	// 去拉下载文件
	resp, err = http.Get(DownloadURL)
	if err != nil {
		return err
	}
	bin, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fw, err := w.CreateFormFile("file", "")
	if err != nil {
		return err
	}
	_, err = fw.Write(bin)
	if err != nil {
		return err
	}
	w.Close()
	// 发起第二次Upload POST请求
	req, err := http.NewRequest("POST", rsp.URL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Read response failed.")
	}
	return nil
}
func main() {
	// 获取旧的小程序的access_token
	OldAccessToken := GetAccessToken(OldAppID, OldAppSecret)
	// 获取新的小程序的access_token
	NewAccessToken := GetAccessToken(NewAppID, NewAppSecret)
	// 已获取，启用转移路由
	http.HandleFunc("/file-transfer", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// 获取路径
		path := r.FormValue("path")
		fileID := r.FormValue("id")
		if path == "" || fileID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// 开始获取下载路径
		DownloadURL, err := GetDownloadURL(OldAccessToken, fileID)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		fmt.Printf("已收到云储存文件回复：\n文件路径：%s\n下载URL：%s\n", path, DownloadURL)
		// 调用API上传
		fmt.Printf("正在对文件路径：%s，下载URL：%s文件发起上传请求\n", path, DownloadURL)
		err = UploadFile(NewAccessToken, path, DownloadURL)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	})
	http.ListenAndServe("127.0.0.1:20000", nil)
}
