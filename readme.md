## better-qiniu

> 封装七牛 Go SDK，以更友好的接口上传文件 

```go
package main 
import (
    qn "github.com/og/go-better-qiniu"
	"github.com/qiniu/api.v7/v7/storage"
	"time"
)
qiniuClient := qn.Client{
    AK: TestAK,
    SK: TestSK,
    PutPolicy: storage.PutPolicy{
        Scope: "og-demo", // 空间名
    },
    StorageConfig: storage.Config{
        Zone:          &storage.ZoneHuanan,
    },
}
resp, err := qiniuClient.Upload(qn.Upload{
    LocalFilename: "localfile.txt",
    QiniuFilename: "name.txt",
    PutExtra:      storage.PutExtra{},
}) ; if err != nil {panic(err)}
// 公开空间
qiniuClient.PublicURL("http://domain.com", resp.Key)
// 私有空间
qiniuClient.PrivateURL("http://domain.com", resp.Key, time.Minute*10)

// 分片上传大文件
qiniuClient.ResumeUpload(qn.ResumeUpload{
    LocalFilename: "localfile.text",
    QiniuFilename: "name.txt",
    RputExtra:     storage.RputExtra{},
})

// 直接上传少了字节，大量文件建议 分批读取通过 file os.O_APPEND 插入本地文件后使用 ResumeUpload 上传
qiniuClient.BytesUpdate(qn.BytesUpdate{
    QiniuFilename: "name.txt",
    Data: []byte("abc"),
    RputExtra:     storage.RputExtra{},
})
```