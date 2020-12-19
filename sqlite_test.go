package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

//TODO: create database just for testing
func TestFindNoArgs(t *testing.T) {
	db, err := NewDB("db/rw.db")
	check(t, err)
	defer db.Close()
	rows, n, err := db.Query("select * from User", nil)
	check(t, err)
	fmt.Println(n)
	for _, row := range rows {
		fmt.Println(row)
	}
}

func TestFindNoArgsNull(t *testing.T) {
	db, err := NewDB("db/rw.db")
	check(t, err)
	defer db.Close()
	rows, n, err := db.Query("select id, image from User", nil)
	check(t, err)
	fmt.Println(n)
	for _, row := range rows {
		fmt.Println(row)
	}
}

func TestFindWithArgs(t *testing.T) {
	db, err := NewDB("db/rw.db")
	check(t, err)
	defer db.Close()
	findTest := func(email string) string {
		rows, n, err := db.Query("select id, email, emailConfirmed from User where email= $email",
			Args{"$email": email})
		check(t, err)
		if n != 1 {
			t.Errorf("wrong result, wanted %d records, got %d", 1, n)
			return "nothing found"
		}
		json, err := json.MarshalIndent(rows[0], "", "")
		check(t, err)
		// if rows[0][1] != "t1@t.ca" {
		// }
		return string(json)
	}
	t.Log(findTest("t1@t.ca"))
	t.Log(findTest("t2@t.ca"))
	t.Log(findTest("t1@t.can"))
	t.Log(findTest("t3@t.ca"))
}

func TestExec(t *testing.T) {
	db, err := NewDB("db/rw.db")
	check(t, err)
	defer db.Close()
	n, id, err := db.Exec("INSERT INTO User(email, userName) Values($email,$username)",
		Args{"$email": "t6@t.ca", "$username": "t6"})
	check(t, err)
	if n != 1 {
		t.Errorf("wrong result, wanted %d records, got %d", 1, n)
	}
	if id == 0 {
		t.Errorf("wrong result, wanted non-zero rowid")
	}
}

// func Test_sqlite_Find(t *testing.T) {
// 	type fields struct {
// 		pool *sqlx.Pool
// 	}
// 	type args struct {
// 		query string
// 		args  Args
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		args    args
// 		want    []interface{}
// 		want1   int
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			db := &sqlite{
// 				pool: tt.fields.pool,
// 			}
// 			got, got1, err := db.Query(tt.args.query, tt.args.args)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("sqlite.Query() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("sqlite.Query() got = %v, want %v", got, tt.want)
// 			}
// 			if got1 != tt.want1 {
// 				t.Errorf("sqlite.Query() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }
