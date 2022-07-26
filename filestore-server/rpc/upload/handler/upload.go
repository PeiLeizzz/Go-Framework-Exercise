package handler

import (
	"context"
	"filestore-server/rpc/upload/config"
	proto "filestore-server/rpc/upload/proto"
)

type Upload struct {
}

func (u *Upload) UploadEntry(ctx context.Context, req *proto.ReqEntry, res *proto.RespEntry) error {
	res.Entry = config.UploadEntry
	return nil
}
