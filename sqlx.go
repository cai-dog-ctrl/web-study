package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)
type users struct {
	ID int
	Age int
	Name string
}
var db *sqlx.DB
//连接数据库
func initDB()(err error){
	dns:="root:lh1296643805@tcp(127.0.0.1)/web?charset=utf8mb4&parseTime=True"
	db,err =sqlx.Connect("mysql",dns)
	if err!=nil{
		fmt.Printf("initDB err:%v\n",err)
		return
	}
	db.SetMaxOpenConns(20)//设置到数据库的最大打开连接数。
	db.SetMaxIdleConns(5)//置空闲连接池中的最大连接数。
	return
}
// 查询单条数据示例
func queryRowDemo() {
	sqlStr := "select id, name, age from users where id=?"
	var u users
	err := db.Get(&u, sqlStr, 1)
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return
	}
	fmt.Printf("id:%d name:%s age:%d\n", u.ID, u.Name, u.Age)
}
func main(){
	err:=initDB()
	if err!=nil{
		fmt.Printf("initDB err ")
	}
}
