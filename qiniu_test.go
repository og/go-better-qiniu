package qn_test

import (
	qn "github.com/og/go-better-qiniu"
	ge "github.com/og/x/error"
	gtest "github.com/og/x/test"
	"github.com/qiniu/api.v7/v7/storage"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func ExampleBasic() {
	qiniuClient := qn.Client{
		AK: TestAK,
		SK: TestSK,
		Bucket: TestBucket,
		StorageConfig: storage.Config{
			Zone:          &storage.ZoneHuanan,
		},
	}
	resp, err := qiniuClient.Upload(qn.Upload{
		LocalFilename: "localfile.txt",
		QiniuFileKey: "name.txt",
	}) ; if err != nil {panic(err)}
	// 公开空间
	qiniuClient.PublicURL("http://domain.com", resp.Key)

	// 分片上传大文件
	qiniuClient.ResumeUpload(qn.ResumeUpload{
		LocalFilename: "localfile.text",
		QiniuFileKey: "name.txt",
		RputExtra:     storage.RputExtra{},
	})

	// 直接上传少了字节，大量文件建议 分批读取通过 file os.O_APPEND 插入本地文件后使用 ResumeUpload 上传
	qiniuClient.BytesUpdate(qn.BytesUpdate{
		QiniuFileKey: "name.txt",
		Data: []byte("abc"),
		RputExtra:     storage.RputExtra{},
	})
}
func TestFile(t *testing.T) {
	as := gtest.NewAS(t)
	qiniuClient := qn.Client{
		AK: TestAK,
		SK: TestSK,
		Bucket: TestBucket,
		StorageConfig: storage.Config{
			Zone:          &storage.ZoneHuanan,
		},
	}
	{
		resp, err := qiniuClient.Upload(qn.Upload{
			LocalFilename: "go.mod",
			QiniuFileKey: time.Now().Format("20060102150405") + "golangmod",
		})
		as.NoError(err, "can not be error")
		log.Print(resp)
	}
	{
		_, err := qiniuClient.ResumeUpload(qn.ResumeUpload{
			LocalFilename: "go.mod",
			QiniuFileKey: time.Now().Format("20060102150405") + "demo.txt",
		})
		as.NoError(err, "can not be error")
	}
	{
		cloudFilename := time.Now().Format("20060102150405") + "byte.txt"
		resp, err := qiniuClient.BytesUpdate(qn.BytesUpdate{
			QiniuFileKey: cloudFilename,
			Data:          []byte("abc"),
			RputExtra:     storage.RputExtra{},
		})
		as.NoError(err, "can not be error")
		as.Equal(resp.Key, cloudFilename)
		url := qiniuClient.PrivateURL(qn.PrivateURLData{
			Domain:   TestDomain,
			Key:      resp.Key,
			Duration: time.Second*10,
			Attname:  "",
		})
		httpResp , err := http.Get(url) ; ge.Check(err)
		data, err := ioutil.ReadAll(httpResp.Body) ;ge.Check(err)
		log.Print(url)
		as.Equal(data, []byte("abc"))
		err = qiniuClient.BucketManager().Delete(TestBucket, resp.Key) ; if err != nil {panic(err)}
	}
	{
		cloudFilename := time.Now().Format("20060102150405") + "byte.txt"
		resp, err := qiniuClient.BytesUpdate(qn.BytesUpdate{
			QiniuFileKey: cloudFilename,
			Data:          []byte("abc"),
			RputExtra:     storage.RputExtra{},
		})
		as.NoError(err, "can not be error")
		as.Equal(resp.Key, cloudFilename)
		url := qiniuClient.PrivateURL(qn.PrivateURLData{
			Domain:   TestDomain,
			Key:      resp.Key,
			Duration: time.Second*10,
			Attname:  time.Now().Format("20060102150405") + "othername.csv",
		})
		log.Print(url)
	}
	{
		source  := []qn.ZipData{}
		{
			one, err := qiniuClient.BytesUpdate(qn.BytesUpdate{
				QiniuFileKey: "zip-source-1.txt",
				Data:         []byte("1"),
			})
			as.NoError(err)
			source = append(source, qn.ZipData{
				QiniuFileKey: one.Key,
				ZipRename: "zip-1.txt",
			})
		}
		{
			two, err := qiniuClient.BytesUpdate(qn.BytesUpdate{
				QiniuFileKey: "zip-source-2.txt",
				Data:         []byte("2"),
			})
			as.NoError(err)
			source = append(source, qn.ZipData{
				QiniuFileKey: two.Key,
				ZipRename: "zip-2.txt",
			})
		}
		persistentID, err := qiniuClient.Pfop(qn.Pfop{
			Source:          source,
			QiniuZipFileKey: "zip/" + time.Now().Format("20060102150405") + "file.zip",
		})
		as.NoError(err)
		status, err := qiniuClient.Prefop(persistentID)
		as.NoError(err)
		log.Printf("pfop status %+v", status)
	}
}
func TestPing(t *testing.T) {
	as := gtest.NewAS(t)
	{
		qiniuClient := qn.Client{
			AK: TestAK,
			SK: TestSK,
			Bucket: TestBucket,
			StorageConfig: storage.Config{
				Zone:          &storage.ZoneHuanan,
			},
		}
		as.NoError(qiniuClient.Ping())
	}

	{
		qiniuClient := qn.Client{
			AK: "",
			SK: TestSK,
			Bucket: TestBucket,
			StorageConfig: storage.Config{
				Zone:          &storage.ZoneHuanan,
			},
		}
		as.ErrorString(qiniuClient.Ping(),"bad token")
	}
	{
		qiniuClient := qn.Client{
			AK: TestAK,
			SK: "",
			Bucket: TestBucket,
			StorageConfig: storage.Config{
				Zone:          &storage.ZoneHuanan,
			},
		}
		as.ErrorString(qiniuClient.Ping(),"bad token")
	}
	{
		qiniuClient := qn.Client{
			AK: TestAK,
			SK: TestSK,
			Bucket: "",
			StorageConfig: storage.Config{
				Zone:          &storage.ZoneHuanan,
			},
		}
		as.ErrorString(qiniuClient.Ping(),"no such bucket")
	}

}

