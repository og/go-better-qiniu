package qn

import (
	"bytes"
	"context"
	"fmt"
	ge "github.com/og/x/error"
	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"os"
	"strings"
	"time"
)
func createCallReader (reader func()(end bool, data []byte), file *os.File) {
	end, data := reader()
	_, err := file.Write(data) ; ge.Check(err)
	if !end {
		createCallReader(reader, file)
	}
}
func Create(filename string, reader func()(end bool, data []byte)) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666) ; ge.Check(err)
	createCallReader(reader, file)
	defer ge.Check(file.Close())
}
type Client struct {
	AK string
	SK string
	PutPolicy storage.PutPolicy
	StorageConfig storage.Config
}

func (q Client) Token() string {
	return q.PutPolicy.UploadToken(q.Mac())
}
func (q Client) Mac() *qbox.Mac {
	return qbox.NewMac(q.AK, q.SK)
}
func (q Client) Credentials() *auth.Credentials {
	return auth.New(q.AK,q.SK)
}
type ResumeUpload struct {
	LocalFilename string
	QiniuFilename string
	RputExtra storage.RputExtra
}
func (q Client) ResumeUpload(data ResumeUpload) (resp Response ,err error) {
	q.PutPolicy.Scope += ":" + data.QiniuFilename
	uploader := storage.NewResumeUploader(&q.StorageConfig)
	err = uploader.PutFile(context.Background(), &resp, q.Token(), data.QiniuFilename, data.LocalFilename, &data.RputExtra)
	return
}
type BytesUpdate struct {
	QiniuFilename string
	Data []byte
	RputExtra storage.RputExtra
}
func (q Client) BytesUpdate(data BytesUpdate)(resp Response ,err error)  {
	q.PutPolicy.Scope += ":" + data.QiniuFilename
	uploader := storage.NewResumeUploader(&q.StorageConfig)
	err = uploader.Put(context.Background(), &resp, q.Token(), data.QiniuFilename, bytes.NewReader(data.Data), int64(len(data.Data)), &data.RputExtra)
	return
}
type Upload struct {
	LocalFilename string
	QiniuFilename string
	PutExtra storage.PutExtra
}
func (q Client) Upload(data Upload) (resp Response ,err error) {
	q.PutPolicy.Scope += ":" + data.QiniuFilename
	uploader := storage.NewFormUploader(&q.StorageConfig)

	err = uploader.PutFile(context.Background(), &resp, q.Token(), data.QiniuFilename, data.LocalFilename, &data.PutExtra)
	return
}
type Response struct {
	Hash         string `json:"hash"`
	PersistentID string `json:"persistentId"`
	Key          string `json:"key"`
}
func (q Client) PublicURL(domain string, key string) string {
	return storage.MakePublicURL(domain, key)
}
type PrivateURLData struct {
	Domain string
	Key string
	Duration time.Duration
	Attname string
}
func (q Client) PrivateURL(data PrivateURLData) string {
	publicURL := q.PublicURL(data.Domain, data.Key)
	urlToSign := publicURL
	if strings.Contains(publicURL, "?") {
		urlToSign = fmt.Sprintf("%s&e=%d", urlToSign, data.Duration)
	} else {
		urlToSign = fmt.Sprintf("%s?e=%d", urlToSign, data.Duration)
	}
	if len(data.Attname) != 0 {
		urlToSign += "&attname=" + data.Attname
	}
	token := q.Credentials().Sign([]byte(urlToSign))
	privateURL := fmt.Sprintf("%s&token=%s", urlToSign, token)
	return privateURL
}
func (q Client) BucketManager () *storage.BucketManager {
	return storage.NewBucketManager(q.Credentials(), &q.StorageConfig)
}
func (q Client) BucketName() string {
	return q.PutPolicy.Scope
}
func (q Client) Ping () error {
	err := q.BucketManager().DeleteAfterDays(q.BucketName(), "Nonexistentfile__0102012", 0)
	if err.Error() == "no such file or directory" {
		return nil
	}
	return err
}