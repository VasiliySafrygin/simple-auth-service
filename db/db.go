package db

import (
	"auth-service/function"
	"auth-service/models"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type DB struct {
	db *sql.DB
}

var connection *sql.DB

// func Database(connectionString string) gin.HandlerFunc {
// 	if connection == nil {
// 		fmt.Fprintf(os.Stderr, "Connected to database ...")
// 		var err error
// // 		connection, err = sql.Open("postgres", connectionString)
//         connection, err = sql.Open("voltdb", connectionString)
// 		if err != nil {
// 			log.Panic(err)
// 		}
// 	} else {
// 		fmt.Fprintf(os.Stderr, "Connected to database: %v\n", connection)
// 	}

// 	return func(c *gin.Context) {
// 		c.Set("DB", &DB{connection})
// 		c.Next()
// 	}
// }

func Connect(connectionString string) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			fmt.Printf("%v, %s", panicInfo, string(debug.Stack()))
		}
	}()

	var err error

	connection, err = sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "Connected to database: %v\n", connection)
}

func GetGroups() ([]models.Group, error) {
	groups := []models.Group{}
	err := connection.Ping()
	if err != nil {
		connection.Close()
		return nil, err
	}
	rows, err := connection.Query("SELECT id, name FROM auth_group")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var group models.Group
		err := rows.Scan(&group.Id, &group.Name)
		groups = append(groups, group)
		if err != nil {
			log.Fatal(err)
		}
	}
	return groups, nil
}

func GetUserByName(name string) (models.User, error) {
	err := connection.Ping()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unableconnect to db: %v\n", err)
		connection.Close()
		return models.User{}, err
	}
	rows, err := connection.Query("SELECT id, username, first_name, last_name, middle_name, password FROM users where username = $1 limit 1", name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		log.Fatal(err)
	}

	defer rows.Close()

	user := models.User{}
	for rows.Next() {
		s := reflect.ValueOf(&user).Elem()
		numCols := s.NumField()
		columns := make([]interface{}, numCols)
		for i := 0; i < numCols; i++ {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		err := rows.Scan(columns...)
		if err != nil {
			log.Fatal(err)
		}
	}
	//for rows.Next() {
	//	err := rows.Scan(&user.Id, &user.UserName, &user.FirstName, &user.LastName, &user.MiddleName, &user.Password)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}

	return user, nil
}

func GetUserById(id string) (models.User, error) {
	err := connection.Ping()
	if err != nil {
		connection.Close()
		return models.User{}, err
	}
	rows, err := connection.Query("SELECT id, username, first_name, last_name, middle_name, password FROM users where id = $1", id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		sentry.CaptureException(err)
		log.Fatal(err)
	}

	defer rows.Close()
	var user models.User
	for rows.Next() {
		err := rows.Scan(&user.Id, &user.UserName, &user.FirstName, &user.LastName, &user.MiddleName, &user.Password)
		if err != nil {
			log.Fatal(err)
		}
	}
	return user, nil
}

func DeleteToken(id string) error {
	err := connection.Ping()
	if err != nil {
		connection.Close()
		return err
	}
	_, err = connection.Query("DELETE FROM tokens where user_id = $1", id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		log.Fatal(err)
		return err
	}

	return nil
}

func CreateJwtToken(userId string) (string, error) {
	user, err := GetUserById(userId)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wrong userId: %v\n", err)
		log.Fatal(err)
		return "", err
	}
	token, err := function.CreateJwtToken(user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable create token: %v\n", err)
		log.Fatal(err)
		return "", err
	}
	_, err = connection.Query("UPDATE tokens set token = $1 WHERE user_id = $2", token, userId)
	if err != nil {
		_, err = connection.Query("INSERT INTO tokens values ($1, $2)", userId, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
			log.Fatal(err)
			return "", err
		}
	}
	//
	//_, err = connection.Query("INSERT INTO tokens values ($1, $2)", userId, token)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
	//	log.Fatal(err)
	//	return "", err
	//}
	return token, nil
}

func GetToken(id string) (string, error) {
	err := connection.Ping()
	if err != nil {
		connection.Close()
		return "", err
	}
	rows, err := connection.Query("SELECT token FROM tokens where user_id = $1", id)
	defer rows.Close()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		log.Fatal(err)
	}

	var token string
	for rows.Next() {
		err := rows.Scan(&token)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
	}
	if len(token) == 0 {
		token, err = CreateJwtToken(id)
		if err != nil {
			return "", nil
		}
	}

	user_id, err := function.VerifyToken(token)
	if user_id != id {
		token, err = CreateJwtToken(id)
		if err != nil {
			return "", nil
		}
	}

	return token, nil
}

func CreateUser(username string, firstName string, lastName string, middleName string, password string) (models.User, error) {
	defer sentry.Recover()
	err := connection.Ping()
	if err != nil {
		connection.Close()
		return models.User{}, err
	}
	password, err = function.HashPassword(password)
	if err != nil {
		fmt.Println(err)
		return models.User{}, err
	}
	user := models.User{uuid.NewString(), username, firstName, lastName, middleName, password}

	_, err = connection.Query("INSERT INTO users values ($1, $2, $3, $4, $5, $6)", user.Id, user.UserName, user.FirstName, user.LastName, user.MiddleName, user.Password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable execute query: %v\n", err)
		return models.User{}, err
	}

	return user, nil
}
