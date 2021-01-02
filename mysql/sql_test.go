package mysql

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/log"
	"github.com/stretchr/testify/require"
)

type StuInfo struct {
	Id       int64   `ddb:"id"`
	Name     string  `ddb:"name"`
	CourseId int     `ddb:"course_id"`
	Score    float64 `ddb:"score"`
}

func TestSql(t *testing.T) {
	require.Nil(t, conf.Parse("../test/configs"))
	log.Init()
	InitMysql(ModeFile, "motor_test", "default", "test")

	// create table
	dbname := "test"
	tmpDb, err := _mysqlPool.getDB(dbname, WRITE)
	require.Nil(t, err)
	tableName := "test_stu_" + strconv.Itoa(rand.Intn(64))
	createTable(tmpDb, tableName, t)
	defer dropTable(tmpDb, tableName, t)

	db := GetDB(nil, dbname, tableName, WRITE)
	require.NotNil(t, db)

	// select empty table
	where := map[string]interface{}{
		"course_id": 1,
		"score >":   80,
	}
	selectFiles := []string{"id", "name"}
	var sinfo1 []StuInfo
	err = db.GetList(where, selectFiles, &sinfo1)
	fmt.Printf("%+v", err)
	require.Nil(t, err)
	require.Equal(t, len(sinfo1), 0)

	// insert
	var data []map[string]interface{}
	data = append(data, map[string]interface{}{
		"name":      "张三",
		"course_id": 1,
		"score":     85.5,
	},
		map[string]interface{}{
			"name":      "李四",
			"course_id": 1,
			"score":     59,
		},
	)
	_, err = db.Insert(data)
	require.Nil(t, err)

	// select mutil rows
	where = map[string]interface{}{
		"course_id": 1,
		"score >=":  59,
	}
	var sinfo2 []StuInfo
	err = db.GetList(where, nil, &sinfo2)
	require.Nil(t, err)
	require.Equal(t, len(sinfo2), 2)
	require.Equal(t, sinfo2[0].Name, "张三")
	require.Equal(t, sinfo2[0].CourseId, 1)
	require.Equal(t, sinfo2[0].Score, 85.5)
	require.Equal(t, sinfo2[1].Name, "李四")
	require.Equal(t, sinfo2[1].CourseId, 1)
	require.Equal(t, sinfo2[1].Score, float64(59))

	// select single row
	where = map[string]interface{}{
		"score >": 0,
	}
	var sinfo3 StuInfo
	err = db.GetRow(where, nil, &sinfo3)
	require.Nil(t, err)
	require.Equal(t, sinfo3.Name, "张三")

	// update
	where = map[string]interface{}{
		"score <": 60,
	}
	update := map[string]interface{}{
		"score": 60,
	}
	arows, err := db.Update(where, update)
	require.Nil(t, err)
	require.Equal(t, arows, int64(1))

	// custom sql
	sql := "select * from " + tableName + " where course_id = {{courseId}} AND score = {{score}} order by course_id, score"
	da := map[string]interface{}{
		"courseId": 1,
		"score":    60,
	}
	var sinfo4 []StuInfo
	err = db.NamedQuery(sql, da, &sinfo4)
	require.Nil(t, err)
	require.Equal(t, len(sinfo4), 1)
}

func createTable(db *sql.DB, tableName string, t *testing.T) {
	dropTable(db, tableName, t)

	query := "CREATE TABLE " + tableName + `( 
    id bigint NOT NULL AUTO_INCREMENT,
    name varchar(32) NOT NULL,
    course_id int unsigned NOT NULL,
    score float NOT NULL DEFAULT 0.0,
    PRIMARY KEY(id)
	)Engine=InnoDB Charset=utf8;`

	_, err := db.Exec(query)
	require.Nil(t, err)
}

func dropTable(db *sql.DB, tableName string, t *testing.T) {
	query := "DROP TABLE IF EXISTS " + tableName + ";"
	_, err := db.Exec(query)
	require.Nil(t, err)
}

// 手动修改mysql.toml中最大连接数来测试连接池
//func TestPoolConfChange(t *testing.T) {
//	require.Nil(t, conf.Parse("../test/configs"))
//	log.Init()
//	InitMysql(ModeFile, "motor_test", "default", "test")
//
//	// create table
//	dbname := "test"
//	tmpDb, err := _mysqlPool.getDB(dbname, WRITE)
//	require.Nil(t, err)
//	tableName := "test2_stu_" + strconv.Itoa(rand.Intn(64))
//	createTable(tmpDb, tableName, t)
//	defer dropTable(tmpDb, tableName, t)
//
//	db := GetDB(nil, dbname, tableName, WRITE)
//	require.NotNil(t, db)
//
//	// insert
//	var data []map[string]interface{}
//	data = append(data, map[string]interface{}{
//		"name":      "张三",
//		"course_id": 1,
//		"score":     85.5,
//	},
//		map[string]interface{}{
//			"name":      "李四",
//			"course_id": 1,
//			"score":     59,
//		},
//	)
//	_, err = db.Insert(data)
//	require.Nil(t, err)
//
//	go func() {
//		time.Sleep(20*time.Second)
//		db := GetDB(nil, dbname, tableName, WRITE)
//		require.NotNil(t, db)
//
//		_, err := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE course_id=1", tableName))
//		fmt.Println("2", err)
//		//rows.Close()
//		_, err = db.Query(fmt.Sprintf("SELECT * FROM %s WHERE course_id=1", tableName))
//		fmt.Println("2", err)
//	}()
//	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE course_id=1", tableName))
//	//rows.Close()
//	fmt.Println("1", err)
//	rows2, err := db.Query(fmt.Sprintf("SELECT * FROM %s WHERE course_id=1", tableName))
//	//rows.Close()
//	fmt.Println("1", err)
//}
