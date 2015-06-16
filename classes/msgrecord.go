package streamtcp

/*
import (
	"fmt"
)
*/
const (
	MSG_REQ = iota
	MSG_REP
	MSG_WRT
)

const (
	T_P2P = iota
	T_UAS
	T_SAU
	T_PUB
	T_PUBACK
	T_SUB
	T_SUBACK
	T_UNSUB
	T_UNSUBACK
	T_REG
)

type MsgRequest struct {
	id      string
	ver     string
	msgtype int32
	query   map[string]interface{}
	userid  string
	token   string
}

type MsgResObj struct {
	id     string
	ver    string
	err    int32
	result map[string]interface{}
}

type MsgResArray struct {
	id     string
	ver    string
	err    int32
	result []map[string]interface{}
}

type midQueryObj struct {
	query  chan string
	result chan string
}

type midQueryArray struct {
	query  chan string
	result chan []string
}
