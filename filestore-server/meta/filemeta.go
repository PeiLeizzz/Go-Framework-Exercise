package meta

import (
	"filestore-server/db"
	"sort"
	"strings"
	"sync"
)

// FileMeta: 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string // 存储在本地的路径
	UploadAt string
}

var (
	fileMetas map[string]*FileMeta // sha1: fileMeta
	mu        sync.RWMutex
)

func init() {
	fileMetas = make(map[string]*FileMeta)
}

// NOT USE: UpdateFileMeta: 新增/更新文件元信息
func UpdateFileMeta(fmeta *FileMeta) {
	mu.Lock()
	defer mu.Unlock()
	fileMetas[fmeta.FileSha1] = fmeta
}

// InsertFileMetaDB: 新增文件元信息到数据库中
func InsertFileMetaDB(fmeta *FileMeta) (bool, error) {
	return db.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// UpdateFileMetaDB: 更新文件元信息到数据库中
func UpdateFileMetaDB(fmeta *FileMeta) (bool, error) {
	return db.UpdateFileMeta(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// NOT USE: GetFileMeta: 通过 sha1 获取文件的元信息对象
func GetFileMeta(sha1 string) *FileMeta {
	mu.RLock()
	mu.RUnlock()
	return fileMetas[sha1]
}

// NOT USE: GetLastFileMetas: 获取批量的文件元信息列表
func GetLastFileMetas(count int) []*FileMeta {
	fMetas := make([]*FileMeta, 0, count)

	mu.RLock()
	for _, v := range fileMetas {
		fMetas = append(fMetas, v)
	}
	mu.RUnlock()

	sort.Slice(fMetas, func(i, j int) bool {
		return strings.Compare(fMetas[i].UploadAt, fMetas[j].UploadAt) == 1
	})
	return fMetas
}

// GetFileMetaDB: 通过 sha1 从数据库中获取文件的元信息对象
func GetFileMetaDB(sha1 string) (*FileMeta, error) {
	tfile, err := db.GetFileMeta(sha1)
	if err != nil {
		return nil, err
	}

	fmeta := &FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fmeta, nil
}

// RemoveFileMeta: 删除文件元信息
func RemoveFileMeta(fileSha1 string) {
	mu.Lock()
	defer mu.Unlock()
	delete(fileMetas, fileSha1)
}
