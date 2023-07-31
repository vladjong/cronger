package main

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/vladjong/cronger"
)

const (
	dns = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
)

type User struct {
	cr   *cronger.Cronger
	Name string
}

type Type struct {
	Age    int
	Number int
}

func (u *User) Test(job cronger.Job, t Type) {
	if err := u.cr.Add(
		cronger.Fields{
			Job: job,
			Task: func() error {
				return Test("test", t)
			},
		}); err != nil {
		log.Fatalln(err)
	}
}

func Test(name string, t Type) error {
	fmt.Printf("Running %s\n", name)
	fmt.Printf("Type: %v", t)
	return nil
}

func Invoke(any interface{}, name string, job cronger.Job) {
	inputs := make([]reflect.Value, len(job.FunctionFields)+1)
	inputs[0] = reflect.ValueOf(job)
	for i := 1; i < len(inputs); i++ {
		if i == len(inputs)-1 {
			m := job.FunctionFields[i-1].(map[string]interface{})
			t := Type{
				Age:    int(m["Age"].(float64)),
				Number: int(m["Number"].(float64)),
			}
			inputs[i] = reflect.ValueOf(t)
			break
		}
		inputs[i] = reflect.ValueOf(job.FunctionFields[i-1])
	}
	reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

func main() {
	db, err := sqlx.Connect("postgres", dns)
	if err != nil {
		log.Fatalln(err)
	}

	cr, err := cronger.New(
		&cronger.Config{
			Loc:        time.UTC,
			Repository: cronger.NewSqlx(db),
		})
	if err != nil {
		log.Fatalln(err)
	}

	u := User{
		Name: "Nick",
		cr:   cr,
	}

	// Recover Task
	jobs := cr.SuspendJobs()
	fmt.Println(jobs)
	if len(jobs) > 0 {
		for _, job := range jobs {
			Invoke(&u, job.FunctionName, job)
		}
	}

	t := Type{
		Age:    12,
		Number: 11,
	}
	tag1 := uuid.NewString()
	id := uuid.NewString()
	data := []interface{}{
		u.Name,
		t,
	}

	job := cronger.Job{
		Tag:            tag1,
		ID:             id,
		Limit:          1,
		Expression:     "* * * * *",
		FunctionName:   "Test",
		FunctionFields: data,
	}
	if err := cr.Add(
		cronger.Fields{
			Job: job,
			Task: func() error {
				return Test(u.Name, t)
			},
		}); err != nil {
		log.Fatalln(err)
	}

	jobs, err = cr.Jobs()
	if err != nil {
		log.Fatalln(jobs)
	}
	fmt.Println("All jobs:", jobs)

	time.Sleep(time.Minute * 5)
}
