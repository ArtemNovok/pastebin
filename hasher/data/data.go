package data

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/shomali11/util/xhashes"
)

var db *sql.DB

func NewDb(base *sql.DB) {
	db = base
}

var rediscl *redis.Client

func NewRedisClient(cl *redis.Client) {
	rediscl = cl
}

func CreateTable() {
	query := `create table if not exists list(
		id serial primary key,
		ch varchar(20) 
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

}

func GetNum() (int64, error) {
	num, err := CheckDB()
	if err != nil {
		return -1, err
	}
	query := "insert into list (ch) values($1)"
	_, err = db.Exec(query, "1")
	if err != nil {
		return -1, err
	}
	return num, nil
}

func CheckDB() (int64, error) {
	var count int64
	query := `select count(id) from list`
	res, err := db.Query(query)
	if err != nil {
		return -1, err
	}
	defer res.Close()
	for res.Next() {
		err := res.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func GetDelKey() (string, error) {
	ctx := context.Background()
	key, err := GetRandomKey()
	if err != nil {
		return "", err
	}
	res, err := rediscl.GetDel(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}

func GetRandomKey() (string, error) {
	ctx := context.Background()
	key, err := rediscl.RandomKey(ctx).Result()
	if err != nil {
		return "", nil
	}
	return key, err
}

func GetDBSize() (int64, error) {
	ctx := context.Background()
	size, err := rediscl.DBSize(ctx).Result()
	if err != nil {
		return -1, err
	}
	return size, err
}

func GenerateHashes() {
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		num, err := GetNum()
		if err != nil {
			log.Panic(err)
			break
		}
		key := strconv.FormatUint(uint64(xhashes.FNV32(string(num))), 10)
		err = rediscl.Set(ctx, key, key, 0).Err()
		if err != nil {
			log.Panic(err)
			break
		}

	}
}
