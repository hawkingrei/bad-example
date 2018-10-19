// Code generated by protoc-gen-liverpc v0.1, DO NOT EDIT.
// source: v2/Room.proto

/*
Package v2 is a generated liverpc stub package.
This code was generated with go-common/app/tool/liverpc/protoc-gen-liverpc v0.1.

It is generated from these files:
	v2/Room.proto
*/
package v2

import context "context"

import proto "github.com/golang/protobuf/proto"
import "go-common/library/net/rpc/liverpc"

var _ proto.Message // generate to suppress unused imports
// Imports only used by utility functions:

// ==============
// Room Interface
// ==============

type Room interface {
	// * 根据房间id获取房间信息v2
	// 修正：原来的get_info_by_id 在传了fields字段但是不包含roomid的情况下 依然会返回所有字段， 新版修正这个问题， 只会返回指定的字段.
	GetByIds(context.Context, *RoomGetByIdsReq) (*RoomGetByIdsResp, error)
}

// ====================
// Room Live Rpc Client
// ====================

type roomRpcClient struct {
	client *liverpc.Client
}

// NewRoomRpcClient creates a Rpc client that implements the Room interface.
// It communicates using Rpc and can be configured with a custom HTTPClient.
func NewRoomRpcClient(client *liverpc.Client) Room {
	return &roomRpcClient{
		client: client,
	}
}

func (c *roomRpcClient) GetByIds(ctx context.Context, in *RoomGetByIdsReq) (*RoomGetByIdsResp, error) {
	out := new(RoomGetByIdsResp)
	err := doRpcRequest(ctx, c.client, 2, "Room.get_by_ids", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// =====
// Utils
// =====

func doRpcRequest(ctx context.Context, client *liverpc.Client, version int, method string, in, out proto.Message) (err error) {
	err = client.Call(ctx, version, method, in, out)
	return
}