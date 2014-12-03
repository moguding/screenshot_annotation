package main

const REDIS_SCREENSHOT_COMMENT_ZSET_KEY = "sreenshot:comment:zset"
const REDIS_SCREENSHOT_COMMENT_LIST_KEY = "sreenshot:comment:list:IMGID"

const (
	STATUS_OK                      = 0
	STATUS_UNKOWN_ERROR            = 1000
	STATUS_PARAMTERS_ERROR         = 1001
	STATUS_LOGIN_ERROR             = 1002
	STATUS_IMGID_ERROR             = 1003
	STATUS_INVAILD_MESSAGE_ERROR   = 1004
	STATUS_PARSE_MESSAGE_ERROR     = 1005
	STATUS_MESSAGE_TYPE_ERROR      = 1006
	STATUS_MESSAGE_PARAMTERS_ERROR = 1007
)

type Request struct {
	Type    string `json:"type"`
	ImgId   int    `json:"img_id"`
	Uid     int    `json:"uid"`
	PosX    int    `json:"pos_x"`
	PosY    int    `json:"pos_y"`
	Message string `json:"message"`
	UniqId  string `json:"uniqid"`
	Color   string `json:"color"`
}

type Response struct {
	Type    string `json:"type"`
	UniqId  string `json:"uniqid"`
	ImgId   int    `json:"img_id"`
	Uid     int    `json:"uid"`
	PosX    int    `json:"pos_x"`
	PosY    int    `json:"pos_y"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

type CreatePoint struct {
	Type    string `json:"type"`
	UniqId  string `json:"uniqid"`
	ImgId   int    `json:"img_id"`
	Uid     int    `json:"uid"`
	PosX    int    `json:"pos_x"`
	PosY    int    `json:"pos_y"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
	Color   string `json:"color"`
}

type JsonString struct {
	Request
	Time   int64
	UniqId string
}

type Delete struct {
	Type   string `json:"type"`
	ImgId  int    `json:"img_id"`
	Uid    int    `json:"uid"`
	UniqId string `json:"uniqid"`
	PosX   int    `json:"pos_x"`
	PosY   int    `json:"pos_y"`
}

type Color struct {
	Type  string `json:"type"`
	ImgId int    `json:"img_id"`
	Uid   int    `json:"uid"`
	PosX  int    `json:"pos_x"`
	PosY  int    `json:"pos_y"`
	Color string `json:"color"`
}

/*
type Login struct {
	Type string `json:"type"`
	User
}
*/

type MemberInfo struct {
	Type string `json:"type"`
	User []*User
}

type Err struct {
	Type      string `json:"type"`
	ErrorCode int    `json:"error_code"`
}

type User struct {
	Uid      int    `json:"uid"`
	Fullname string `json:"fullname"`
	Mail     string `json:"mail"`
	Icon     string `json:"icon"`
	Status   int    `json:"status"`
}

type CommentMeta struct {
	ImgId int `db:"img_id"`
	//MetaId     string `db:"meta_id"`
	PosX       int    `db:"pos_x"`
	PosY       int    `db:"pos_y"`
	Color      string `db:"color"`
	UpdateTime int64  `db:"update_time"`
}

type CommentInfo struct {
	Uid    int    `db:"uid"`
	UniqId string `db:"uniqid"`
	//MetaId     string `db:"meta_id"`
	Content    string `db:"content"`
	CreateTime int64  `db:"create_time"`
}

type Comment struct {
	CommentMeta *CommentMeta
	CommentInfo []*CommentInfo
	UserInfo    []*User
}

type ResponseResult struct {
	ErrorCode int
	ErrorMsg  string
}
