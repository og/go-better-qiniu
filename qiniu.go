package qn

import (
	"bytes"
	"context"
	ge "github.com/og/x/error"
	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"os"
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
func (q Client) PrivateURL(domain string, key string, duration time.Duration) string {

	return storage.MakePrivateURL(q.Credentials(), domain, key, time.Now().Add(duration).Unix())
}

func (q Client) BucketManager () *storage.BucketManager {
	return storage.NewBucketManager(q.Credentials(), &q.StorageConfig)
}