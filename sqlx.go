package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type users struct {
	ID   int
	Age  int
	Name string
}

func (u users) Value() (driver.Value, error) {
	return []interface{}{u.Name, u.Age}, nil
}

var db *sqlx.DB

//连接数据库
func initDB() (err error) {
	dns := "root:lh1296643805@tcp(127.0.0.1)/web?charset=utf8mb4&parseTime=True"
	db, err = sqlx.Connect("mysql", dns)
	if err != nil {
		fmt.Printf("initDB err:%v\n", err)
		return
	}
	db.SetMaxOpenConns(20) //设置到数据库的最大打开连接数。
	db.SetMaxIdleConns(5)  //置空闲连接池中的最大连接数。
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

// 查询多条数据示例
func queryMultiRowDemo() {
	sqlStr := "select id, name, age from users where id > ?"
	var users []users
	err := db.Select(&users, sqlStr, 0)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return
	}
	fmt.Printf("users:%#v\n", users)
}

//插入数据
func insert() {
	sqlStr := "insert into users(name,age) values (?,?)"
	res, err := db.Exec(sqlStr, "lh2", 17)
	if err != nil {
		fmt.Printf("insert err : %v\n", err)
		return
	}
	theID, err := res.LastInsertId() //新插入的数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d.\n", theID)
}

//更新数据
func update() {
	sqlStr := "update users set name=?,age=? where id =?"
	res, err := db.Exec(sqlStr, "sb", 12, 2)
	if err != nil {
		fmt.Printf("updata err: %v", err)
		return
	}
	n, err := res.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("update success, affected rows:%d\n", n)
}
func delete() {
	sqlStr := "delete from users where id= ?"
	res, err := db.Exec(sqlStr, 2)
	if err != nil {
		fmt.Printf("delete err : %v", err)
		return
	}
	n, err := res.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}

//DB.NamedExec方法用来绑定SQL语句与结构体或map中的同名字段。
func insertUserDemo() (err error) {
	sqlStr := "INSERT INTO users (name,age) VALUES (:name,:age)"
	_, err = db.NamedExec(sqlStr,
		map[string]interface{}{
			"name": "七米",
			"age":  28,
		})
	return
}

//NamedQuery
func namedQuery() {
	sqlStr := "SELECT * FROM users WHERE name=:name"
	rows, err := db.NamedQuery(sqlStr, map[string]interface{}{"name": "七米"})
	if err != nil {
		fmt.Printf("nameQuery err :%v", err)
		return
	}
	for rows.Next() {
		var u users
		err := rows.StructScan(&u)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			continue
		}
		fmt.Printf("user:%#v\n", u)
	}
	u := users{
		Name: "七米",
	}
	// 使用结构体命名查询，根据结构体字段的 db tag进行映射
	rows, err = db.NamedQuery(sqlStr, u)
	if err != nil {
		fmt.Printf("db.NamedQuery failed, err:%v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var u users
		err := rows.StructScan(&u)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			continue
		}
		fmt.Printf("user:%#v\n", u)
	}
}

//事务
func transactionDemo2() (err error) {
	tx, err := db.Beginx() // 开启事务
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	sqlStr1 := "Update users set age=20 where id=?"

	rs, err := tx.Exec(sqlStr1, 1)
	if err != nil {
		return err
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	sqlStr2 := "Update users set age=50 where id=?"
	rs, err = tx.Exec(sqlStr2, 5)
	if err != nil {
		return err
	}
	n, err = rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	return err
}

//使用sqlx.In批量插入数据
func batchInsertUsers(users []interface{}) error {
	sqlStr := "insert users (name, age) values (?) ,(?),(?)"
	query, args, _ := sqlx.In(sqlStr, users...)
	fmt.Println(query)
	fmt.Println(args)
	_, err := db.Exec(query)
	return err
}

// BatchInsertUsers3 使用NamedExec实现批量插入
func BatchInsertUsers3(users []*users) error {
	_, err := db.NamedExec("INSERT INTO users (name, age) VALUES (:name, :age)", users)
	return err
}

// QueryByIDs 根据给定ID查询
func QueryByIDs(ids []int) (users []users, err error) {
	// 动态填充id
	query, args, err := sqlx.In("SELECT name, age FROM users WHERE id IN (?)", ids)
	if err != nil {
		return
	}
	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
	query = db.Rebind(query)

	err = db.Select(&users, query, args...)
	return
}
func main() {
	err := initDB()
	if err != nil {
		fmt.Printf("initDB err ")
	}
	namedQuery()

}
