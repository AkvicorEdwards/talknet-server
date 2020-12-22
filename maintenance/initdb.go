package maintenance

import (
	"log"
	"os"
	"talknet/dam"
)

const (
	sql = `
create table inc
(
    name text,
    val  integer
);

` + `

create table user
(
    uuid       integer
        constraint user_pk
            primary key,
    username   text,
    nickname   text,
    password   text,
    friends    text,
    groups     text,
    unaccepted text,
    deleted    integer default 0
);

create unique index user_username_uindex
    on user (username);
` + `
create table groups
(
    guid       integer
        constraint groups_pk
            primary key,
    name       text,
    owner      integer,
    admin      text,
    member     text,
    unaccepted text,
    files      text,
    deleted    integer default 0
);

create unique index groups_name_uindex
    on groups (name);
` + `

INSERT INTO inc (name, val) VALUES ('user', 0);
INSERT INTO inc (name, val) VALUES ('groups', 0);
`
)

func InitDatabase() {
	if !IsFile("talknet.db") {
		log.Println("talknet.db do not exist")
		os.Exit(-1)
	}

	err := dam.Exec(sql).Error
	if err != nil {
		log.Println(err)
		os.Exit(-2)
	}

	log.Println("Finished")

	os.Exit(0)
}

func Exists(path string) bool {
	_, err := os.Stat(path)    //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	if !Exists(path) {
		return false
	}
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}