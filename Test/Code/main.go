package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/boltdb/bolt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type User struct {
	Username string
	Password string
	Email    string
	Phone    string
}

type Article struct {
	Id      string
	Title   string
	Content string
	Author  string
}

type Critic struct {
	C_id    string
	Id      string
	Content string
}

type Tag struct {
	Tagname string
}

func Hello(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFiles("./hello.html")
	if err != nil {
		log.Fatal("template.ParseFiles : ", err)
		return
	}
	t.Execute(w, nil)
}

func Test(w http.ResponseWriter, r *http.Request) {
	user := User{
		Username: "1",
		Password: "123",
		Email:    "2543761065@qq.com",
		Phone:    "1235",
	}
	json.NewEncoder(w).Encode(user)
}

func Index(w http.ResponseWriter, r *http.Request) {

	t, err := ioutil.ReadFile("./index.html")
	if err != nil {
		log.Fatal("template.ParseFiles : ", err)
		return
	}
	x := string(t)

	re3, _ := regexp.Compile("Result:")
	x = re3.ReplaceAllString(x, "fuck")
	//fmt.Println(x)

	fmt.Fprintf(w, string(t))

}

func Register1(w http.ResponseWriter, r *http.Request) {
	t, err := ioutil.ReadFile("./register.html")
	if err != nil {
		log.Fatal("template.ParseFiles : ", err)
		return
	}
	fmt.Fprintf(w, string(t))
}

func Login1(w http.ResponseWriter, r *http.Request) {
	t, err := ioutil.ReadFile("./login.html")
	if err != nil {
		log.Fatal("template.ParseFiles : ", err)
		return
	}
	fmt.Fprintf(w, string(t))
}

//func UserToBytes(s *User) []byte {
//	var x reflect.SliceHeader
//	x.Len = unsafe.Sizeof(*s)
//	x.Cap = unsafe.Sizeof(*s)
//	x.Data = uintptr(unsafe.Pointer(s))
//	return *(*[]byte)(unsafe.Pointer(&x))
//}

func Register2(w http.ResponseWriter, r *http.Request) {
	inputs := mux.Vars(r)
	fmt.Println(inputs)
	db, err := bolt.Open("chaorsBlock.db", 0600, nil)
	exist := false
	err = db.Update(func(tx *bolt.Tx) error {
		fmt.Println("db")
		b := tx.Bucket([]byte("MyBlocks"))
		if b != nil {
			data := b.Get([]byte(inputs["username"]))
			if data != nil {
				exist = true
				fmt.Println("exist")
				m := User{}

				_ = json.Unmarshal(data, &m)
				fmt.Println(m.Username)
				return nil
			}
			fmt.Println("123")
			t := User{
				Username: inputs["username"],
				Password: inputs["password"],
				Email:    inputs["email"],
				Phone:    inputs["phone"],
			}
			tt, err := json.Marshal(t)
			err = b.Put([]byte(inputs["username"]), tt)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})
	db.Close()
	if err != nil {
		fmt.Println("there")
		log.Fatal(err)
	}
	fmt.Println("here")
	if exist == true {
		fmt.Println("no")
		fmt.Fprint(w, "NO")
	} else {
		fmt.Fprint(w, "YES")
	}
}

//func CreateTABLE() {
//	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
//	_ = db.Update(func(tx *bolt.Tx) error {
//
//		//判断要创建的表是否存在
//		b := tx.Bucket([]byte("Articles"))
//		if b == nil {
//
//			//创建叫"MyBucket"的表
//			_, err := tx.CreateBucket([]byte("Articles"))
//			if err != nil {
//				//也可以在这里对表做插入操作
//				log.Fatal(err)
//			}
//		}
//
//		//一定要返回nil
//		return nil
//	})
//}

// 创建Token
func createToken(key string, m map[string]interface{}) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	for index, val := range m {
		claims[index] = val
	}

	token.Claims = claims
	tokenString, _ := token.SignedString([]byte(key))
	return tokenString
}

// 解析Token
func parseToken(tokenString string, key string) (interface{}, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		fmt.Println(err)
		return "", false
	}
}

func Login3(w http.ResponseWriter, r *http.Request) {
	inputs := mux.Vars(r)
	fmt.Println(inputs)
	key := "key"
	_, flag := parseToken(inputs["token"], key)
	if flag {
		fmt.Fprint(w, "YES")
	} else {
		fmt.Fprint(w, "NO")
	}
}

func Login2(w http.ResponseWriter, r *http.Request) {
	inputs := mux.Vars(r)
	fmt.Println(inputs)
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	exist := false
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBlocks"))
		if b != nil {
			data := b.Get([]byte(inputs["username"]))
			if data != nil {
				m := User{}
				_ = json.Unmarshal(data, &m)
				fmt.Println(m)
				if m.Password != inputs["password"] {
					return nil
				}
				exist = true
				return nil
			}
		}
		return nil
	})
	db.Close()
	if exist {
		m := make(map[string]interface{})
		m["username"] = inputs["username"]
		m["password"] = inputs["password"]
		key := "key"
		tokenstring := createToken(key, m)
		fmt.Fprint(w, tokenstring)
	} else {
		fmt.Fprint(w, "NO")
	}
}

func Userinfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("userinfo")
	inputs := mux.Vars(r)
	name, phone, email := "", "", ""
	exist := false
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBlocks"))
		if b != nil {
			data := b.Get([]byte(inputs["username"]))
			if data != nil {
				m := User{}
				_ = json.Unmarshal(data, &m)
				fmt.Println(m)
				if m.Password != inputs["password"] {
					return nil
				}
				name = m.Username
				phone = m.Phone
				email = m.Email
				exist = true
				return nil
			}
		}
		return nil
	})
	db.Close()

	if exist != false {

		t, _ := ioutil.ReadFile("./user.html")
		x := string(t)
		re3, _ := regexp.Compile("#name")
		x = re3.ReplaceAllString(x, name)
		re3, _ = regexp.Compile("#phone")
		x = re3.ReplaceAllString(x, phone)
		re3, _ = regexp.Compile("#email")
		x = re3.ReplaceAllString(x, email)
		fmt.Fprint(w, x)
	}
	return
	fmt.Fprintf(w, "error")
}

func getAllarticles(w http.ResponseWriter, r *http.Request) {
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	inputs := mux.Vars(r)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Articles"))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf(inputs["username"])
			fmt.Printf("key=%s, value=%s\n", k, v)
			m := Article{}
			_ = json.Unmarshal(v, &m)
			if m.Author == inputs["username"] {
				json.NewEncoder(w).Encode(&m)
			}
		}
		return nil
	})
	db.Close()
}

func getAllcritics(w http.ResponseWriter, r *http.Request) {
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	inputs := mux.Vars(r)
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Critics"))

		if b != nil {
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				m := Critic{}
				_ = json.Unmarshal(v, &m)
				if m.Id == inputs["Id"] {
					json.NewEncoder(w).Encode(&m)
				}
			}
		}

		return nil
	})
	db.Close()
}

var Count int = 0

func CreateCritic(w http.ResponseWriter, r *http.Request) {
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	_ = db.Update(func(tx *bolt.Tx) error {

		//判断要创建的表是否存在
		b := tx.Bucket([]byte("Critics"))
		if b == nil {

			//创建叫"MyBucket"的表
			_, err := tx.CreateBucket([]byte("Critics"))
			if err != nil {
				//也可以在这里对表做插入操作
				log.Fatal(err)
			}
		}

		//一定要返回nil
		return nil
	})
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Critics"))
		if b != nil {
			var critic Critic
			_ = json.NewDecoder(r.Body).Decode(&critic)
			critic.C_id = string(Count)
			Count += 1
			tt, err := json.Marshal(critic)
			err = b.Put([]byte(critic.C_id), tt)
			if err != nil {
				fmt.Fprint(w, "NO")
			} else {
				fmt.Fprint(w, "YES")
			}

		}
		return nil
	})
	db.Close()
}

var Total int = 0

func CreateTag(w http.ResponseWriter, r *http.Request) {
	fmt.Println("11xx")
	inputs := mux.Vars(r)
	exist := false
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyTag"))
		if b != nil {
			data := b.Get([]byte(inputs["tagcontent"]))
			if data != nil {
				exist = true
				return nil
			}
			b.Put([]byte(inputs["tagcontent"]), []byte(inputs["tagcontent"]))
		}
		return nil
	})

	db.Close()

	if exist {
		fmt.Fprint(w, "NO")
	} else {
		fmt.Fprint(w, "YES")
	}

}

func GetTag(w http.ResponseWriter, r *http.Request) {
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyTag"))
		c := b.Cursor()
		fmt.Fprint(w, "TAG有: ")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			fmt.Fprint(w, string(k))
			fmt.Fprint(w, ", ")
		}
		return nil
	})

	db.Close()
}

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Create aiticle")
	//Total := 0
	inputs := mux.Vars(r)
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)

	exist := false
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBlocks"))
		if b != nil {
			data := b.Get([]byte(inputs["username"]))
			if data != nil {
				m := User{}
				_ = json.Unmarshal(data, &m)
				fmt.Println(m)
				if m.Password != inputs["password"] {
					return nil
				}
				exist = true
				return nil
			}
		}
		return nil
	})
	if exist == false {
		fmt.Fprint(w, "error")
		db.Close()
		return
	}

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("Articles"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			v = v
			Total = 1 + Total
			fmt.Printf("null\n")
		}
		return nil
	})

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Articles"))
		var article Article
		_ = json.NewDecoder(r.Body).Decode(&article)
		article.Id = string(Total)
		article.Author = inputs["username"]
		a, _ := json.Marshal(article)
		myerr := b.Put([]byte(string(Total)), a)
		if myerr != nil {
			fmt.Fprint(w, "NO")
		} else {
			fmt.Fprint(w, "YES")
		}
		return nil
	})

	db.Close()
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("get article")
	inputs := mux.Vars(r)
	m := Article{}
	getit := false
	db, _ := bolt.Open("chaorsBlock.db", 0600, nil)
	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Articles"))
		if b != nil {
			fmt.Println([]byte(inputs["Id"]))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("key=%s, value=%s\n", k, v)
				_ = json.Unmarshal(v, &m)
				interg2, _ := strconv.ParseInt(inputs["Id"], 10, 0)
				idbitys := (([]byte)(m.Id))
				if interg2 == (int64)(idbitys[0]) {
					// json.NewEncoder(w).Encode(&m)
					getit = true
					break
				}
			}
		} else {
			fmt.Println("b is nil")
		}
		return nil
	})
	if getit {
		t, _ := ioutil.ReadFile("./article.html")
		x := string(t)
		re3, _ := regexp.Compile("#article_title")
		x = re3.ReplaceAllString(x, m.Title)
		re3, _ = regexp.Compile("#article_content")
		x = re3.ReplaceAllString(x, m.Content)
		re3, _ = regexp.Compile("#article_id")
		x = re3.ReplaceAllString(x, m.Id)
		fmt.Fprint(w, x)
	} else {
		fmt.Println("No getit")
	}
	db.Close()
}

func CreateTABLE() {
	db, err := bolt.Open("chaorsBlock.db", 0600, nil)
	err = db.Update(func(tx *bolt.Tx) error {

		//判断要创建的表是否存在
		b := tx.Bucket([]byte("MyTag"))
		if b == nil {

			//创建叫"MyBucket"的表
			_, err := tx.CreateBucket([]byte("MyTag"))
			if err != nil {
				//也可以在这里对表做插入操作
				log.Fatal(err)
			}
		}

		//一定要返回nil
		return nil
	})

	//更新数据库失败
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("hello world!")
	mymux := mux.NewRouter()
	mymux.HandleFunc("/users", Hello).Methods("GET")
	mymux.HandleFunc("/test", Test).Methods("GET")
	mymux.HandleFunc("/index", Index).Methods("GET")
	mymux.HandleFunc("/register", Register1).Methods("GET")
	mymux.HandleFunc("/register/{username}-{password}-{email}-{phone}", Register2).Methods("POST")
	mymux.HandleFunc("/login/{username}-{password}", Login2).Methods("POST")
	mymux.HandleFunc("/login/{token}", Login3).Methods("POST")
	mymux.HandleFunc("/login", Login1).Methods("GET")
	mymux.HandleFunc("/users/{username}-{password}", Userinfo).Methods("GET")
	mymux.HandleFunc("/articles/{username}", getAllarticles).Methods("GET")
	mymux.HandleFunc("/critics/{Id}", getAllcritics).Methods("GET")
	mymux.HandleFunc("/publish/{username}-{password}", CreateArticle).Methods("POST")
	mymux.HandleFunc("/publish/critic", CreateCritic).Methods("POST")
	mymux.HandleFunc("/detail/{Id}", GetArticle).Methods("GET")
	mymux.HandleFunc("/tag/{tagcontent}", CreateTag).Methods("POST")
	mymux.HandleFunc("/tag", GetTag).Methods("GET")
	fmt.Println("Listning to port 9090")
	err := http.ListenAndServe(":9090", mymux)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
