package filesystem

import (
	"context"
	"github.com/HFO4/cloudreve/models"
	"github.com/HFO4/cloudreve/pkg/filesystem/local"
	"io"
	"path/filepath"
)

// FileData 上传来的文件数据处理器
type FileData interface {
	io.Reader
	io.Closer
	GetSize() uint64
	GetMIMEType() string
	GetFileName() string
}

// Handler 存储策略适配器
type Handler interface {
	Put(ctx context.Context, file io.ReadCloser, dst string) error
}

// FileSystem 管理文件的文件系统
type FileSystem struct {
	/*
	  文件系统所有者
	*/
	User *model.User

	/*
	  钩子函数
	*/
	// 上传文件前
	BeforeUpload func(ctx context.Context, fs *FileSystem, file FileData) error
	// 上传文件后
	AfterUpload func(ctx context.Context, fs *FileSystem) error
	// 文件保存成功，插入数据库验证失败后
	ValidateFailed func(ctx context.Context, fs *FileSystem) error

	/*
	  文件系统处理适配器
	*/
	Handler Handler
}

// NewFileSystem 初始化一个文件系统
func NewFileSystem(user *model.User) (*FileSystem, error) {
	var handler Handler

	// 根据存储策略类型分配适配器
	switch user.Policy.Type {
	case "local":
		handler = local.Handler{}
	default:
		return nil, UnknownPolicyTypeError
	}

	// TODO 分配默认钩子
	return &FileSystem{
		User:    user,
		Handler: handler,
	}, nil
}

// Upload 上传文件
func (fs *FileSystem) Upload(ctx context.Context, file FileData) (err error) {
	// 上传前的钩子
	err = fs.BeforeUpload(ctx, fs, file)
	if err != nil {
		return err
	}

	// 生成文件名和路径
	savePath := fs.GenerateSavePath(file)

	// 保存文件
	err = fs.Handler.Put(ctx, file, savePath)
	if err != nil {
		return err
	}

	return nil
}

// GenerateSavePath 生成要存放文件的路径
func (fs *FileSystem) GenerateSavePath(file FileData) string {
	return filepath.Join(
		fs.User.Policy.GeneratePath(fs.User.Model.ID),
		fs.User.Policy.GenerateFileName(fs.User.Model.ID, file.GetFileName()),
	)
}