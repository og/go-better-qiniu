package qn

import (
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/qiniu/api.v7/v7/storage"
	"time"
)

type ZipData struct {
	QiniuFileKey string
	ZipRename string
}
func (q Client) CreateMkzipIndex(domain string, zips []ZipData, indexFileanme string)(reply Reply ,err error)  {
	var data []byte
	for _, item := range zips {
		url := q.PrivateURL(PrivateURL{
			Key:      item.QiniuFileKey,
			Duration: time.Minute * 120,
		})
		s := "/url/" + base64.StdEncoding.EncodeToString([]byte(url)) +
			"/alias/" + base64.StdEncoding.EncodeToString([]byte(item.ZipRename)) +
			"\n"
		data = append(data, []byte(s)...)
	}
	return q.BytesUpdate(BytesUpdate{
		QiniuFileKey: indexFileanme,
		Data:          data,
		RputExtra:     storage.RputExtra{},
		PutPolicy:     storage.PutPolicy{},
	})
}
type Pfop struct {
	Source []ZipData
	QiniuZipFileKey string
	NotifyURL string
}
func (q Client) Pfop(domain string, data Pfop) (persistentID PersistentID,err error) {
	indexFileKey := "golang/og/go-better-qiniu/mkzip-index/" + uuid.New().String() + ".txt"
	indexReply, err := q.CreateMkzipIndex(domain, data.Source, indexFileKey) ; if err != nil {return "", err}
	om := storage.NewOperationManager(q.Credentials(), &q.StorageConfig)
	key := indexReply.Key
	fops := "mkzip/4/|saveas/" + base64.StdEncoding.EncodeToString([]byte(q.Bucket + ":" + data.QiniuZipFileKey))
	pipeline := ""
	notifyURL := data.NotifyURL
	stringPersistentID, err := om.Pfop(q.Bucket, key, fops, pipeline, notifyURL, false) ; if err != nil {return "", err }
	persistentID = NewPersistentID(stringPersistentID)
	return
}

type PersistentID string
func NewPersistentID (s string) PersistentID {
	return PersistentID(s)
}
func (id PersistentID) String() string {
	return string(id)
}
func (q Client) Prefop(persistentID PersistentID) (ret storage.PrefopRet ,err error) {
	om := storage.NewOperationManager(q.Credentials(), &q.StorageConfig)
	return om.Prefop(persistentID.String())
}