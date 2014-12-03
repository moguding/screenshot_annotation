package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"screenshot/global"
	"screenshot/model"
	"strconv"
	//_ "screenshot/config"
	//_ "screenshot/lib"
	//"strconv"
	//"strings"
	"time"
)

func checkLogin(uid int, session_id string) *global.User {
	m := new(model.UserModel)
	m.Init()
	ret := m.CheckLogin(uid, session_id, false)

	m = nil
	return ret
}

func getUserByUid(uid int) *global.User {
	m := new(model.UserModel)
	m.Init()
	ret := m.GetUserByUid(uid)

	m = nil
	return ret
}

func getImage(img_id int) *global.Image {
	m := new(model.ImageModel)
	m.Init()
	ret := m.GetImageById(int64(img_id))

	m = nil
	return ret
}

func getProjectUser(uid int, img_id int) *global.ProjectUser {
	image := getImage(img_id)

	m := new(model.ProjectUserModel)
	m.Init()
	ret := m.GetProjectUserByUidAndProjectId(int64(image.ProjectId), uid)

	m = nil
	return ret
}

func getProjectAllUser(project_id int64) []global.ProjectUser {
	m := new(model.ProjectUserModel)
	m.Init()
	ret := m.GetAllProjectUserByPorjectId(project_id)

	m = nil
	return ret
}

func getMetaId(img_id, pos_x, pos_y int) string {
	s := strconv.Itoa(img_id) + "#" + strconv.Itoa(pos_x) + "#" + strconv.Itoa(pos_y)
	h := md5.New()
	h.Write([]byte(s))
	s5 := hex.EncodeToString(h.Sum(nil))
	return s5
}

func dbSaveComment(req *JsonString) error {
	var meta_id = getMetaId(req.ImgId, req.PosX, req.PosY)
	//log4w("meta_id,", meta_id, req)
	meta := getCommentMeta(meta_id)

	var sql string

	tx, err := db.Begin()

	if err != nil {
		log4w("db.Begin failed," + err.Error())
		return err
	}

	defer tx.Rollback()

	if req.Color != "" {
		if meta == nil {
			sql = `insert into annotation_meta (meta_id, img_id, pos_x, pos_y, update_time, color)
				values(?, ?, ?, ?, ?, ?)
				`
			_, err = tx.Exec(sql, meta_id, req.ImgId, req.PosX, req.PosY, req.Time, req.Color)
		} else {
			sql = `update annotation_meta set img_id=?, pos_x=?, pos_y=?, update_time=?, color= ?
				where meta_id= ?
				`
			_, err = tx.Exec(sql, req.ImgId, req.PosX, req.PosY, req.Time, req.Color, meta_id)
		}
	} else {
		if meta == nil {
			sql = `insert into annotation_meta (meta_id, img_id, pos_x, pos_y, update_time, color)
				values(?, ?, ?, ?, ?, ?)
				`
			_, err = tx.Exec(sql, meta_id, req.ImgId, req.PosX, req.PosY, req.Time, "")
		} else {
			/*
				sql = `update annotation_meta set img_id=?, pos_x=?, pos_y=?, update_time=?
					where meta_id= ?
					`
			*/
			sql = `update annotation_meta set update_time=?	where meta_id= ? `
			_, err = tx.Exec(sql, req.Time, meta_id)
		}
	}

	if err != nil {
		log4w("tx.Exec failed," + err.Error())
		tx = nil
		return err
	}

	sql = `insert into annotation (uniqid, meta_id, uid, content, create_time)
	values(?, ?, ?, ?, ?)
	`
	_, err = tx.Exec(sql, req.UniqId, meta_id, req.Uid, req.Message, req.Time)
	if err != nil {
		log4w("tx.Exec failed," + err.Error())
		tx.Rollback()
		return err
	} else {
		tx.Commit()
	}

	return nil
}

func getCommentMeta(meta_id string) *CommentMeta {
	var sql string
	sql = "select img_id, pos_x, pos_y, color, update_time from annotation_meta where meta_id= '%s'"
	sql = fmt.Sprintf(sql, meta_id)

	rows, err := db.Query(sql)

	if err != nil {
		log4e(err)
		return nil
	}

	defer rows.Close()

	var img_id int
	var posx int
	var posy int
	var color string
	var update_time int64

	for rows.Next() {
		err = rows.Scan(&img_id, &posx, &posy, &color, &update_time)
		if err == nil {
			meta := CommentMeta{img_id, posx, posy, color, update_time}
			return &meta
		}
	}

	return nil
}

func getNewCommentList(paramter_optional ...int) {
	/*
		var flag= false
		if len(paramter_optional) > 0 {
			flag = paramter_optional[0]
		}
	*/
	var t = time.Now().Unix()
	var sql string

	sql = "select img_id, pos_x, pos_y from annotation_meta where update_time> %d"
	sql = fmt.Sprintf(sql, (t - 1800))

	rows, err := db.Query(sql)

	if err != nil {
		log4e(err)
		return
	}

	defer rows.Close()

	var img_id int
	var pos_x int
	var pos_y int
	log4d(sql)
	for rows.Next() {
		err = rows.Scan(&img_id, &pos_x, &pos_y)
		if err == nil {
			image := getImage(img_id)
			project_user := getProjectAllUser(int64(image.ProjectId))
			log4d(project_user)
			for _, v := range project_user {
				sql := `insert into user_message (uid, img_id, pos_x, pos_y, project_id, company_id, create_time) values
				(?, ?, ?, ?, ?)
				`
				_, _ = db.Exec(sql, v.Uid, img_id, image.ProjectId, image.CompanyId, t)
			}
		}
	}
	return
}

func getCommentList(img_id int64, pos_x int64, pos_y int64, withComment bool) (c []Comment) {
	var sql string
	if pos_x != 0 && pos_y != 0 {
		sql = "select meta_id, pos_x, pos_y, color, update_time from annotation_meta where img_id= %d and pos_x= %d and pos_y= %d"
		sql = fmt.Sprintf(sql, img_id, pos_x, pos_y)
	} else {
		sql = "select meta_id, pos_x, pos_y, color, update_time from annotation_meta where img_id= %d"
		sql = fmt.Sprintf(sql, img_id)
	}

	rows, err := db.Query(sql)

	if err != nil {
		log4e(err)
		return nil
	}

	defer rows.Close()

	var meta_id string
	var posx int
	var posy int
	var color string
	var update_time int64

	for rows.Next() {
		err = rows.Scan(&meta_id, &posx, &posy, &color, &update_time)
		if err == nil {
			meta := CommentMeta{int(img_id), posx, posy, color, update_time}
			if withComment == false {
				c = append(c, Comment{&meta, nil, nil})
			} else {
				info, user := getCommentInfo(meta_id)
				c = append(c, Comment{&meta, info, user})
			}
		}
	}

	return
}

func getCommentInfo(meta_id string) (c []*CommentInfo, u []*User) {
	var sql string

	sql = "select uid, uniqid, content, create_time from annotation where meta_id= '%s' order by id asc"
	sql = fmt.Sprintf(sql, meta_id)

	rows, err := db.Query(sql)

	if err != nil {
		log4e(err)
		return nil, nil
	}

	defer rows.Close()

	var uid int
	var uniqid string
	var content string
	var create_time int64
	var userlist = make(map[int]bool)

	for rows.Next() {
		err = rows.Scan(&uid, &uniqid, &content, &create_time)
		if err == nil {
			info := CommentInfo{uid, uniqid, content, create_time}
			c = append(c, &info)

			if userlist[uid] == true {
				continue
			} else {
				userlist[uid] = true

				user := getUserByUid(uid)
				var icon string
				if user.Icon == "" {
					icon = ""
				} else {
					icon = config.IconUrl + strconv.Itoa(user.Uid)
				}
				userinfo := User{user.Uid, user.Fullname, user.Mail, icon, user.Status}
				u = append(u, &userinfo)
			}
		}
	}

	return
}

func updateCommentMeta(img_id int, pos_x int, pos_y int, color string) error {
	var meta_id = getMetaId(img_id, pos_x, pos_y)
	sql := "update annotation_meta set color= ? where meta_id= ?"
	_, err := db.Exec(sql, color, meta_id)

	if err != nil {
		log4e("updateCommentMeta failed,", err)
		return err
	}

	return nil
}

func deleteComment( /*img_id int, pos_x int, pos_y int, */ uniqid string) error {
	//var meta_id = getMetaId(img_id, pos_x, pos_y)
	sql := "delete from annotation where uniqid= ?"
	_, err := db.Exec(sql, uniqid)

	if err != nil {
		log4e("deleteComment failed,", err)
		return err
	}

	return nil
}
