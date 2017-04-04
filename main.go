// main.go

package main

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Table struct {
	ID			string 	 `form:"id" json:"id" binding:"required"`
	Name    	string   `form:"name" json:"name" binding:"required"`
	Type   		string   `form:"type" json:"type" binding:"required"`
	MinBet 		int32	`form:"min_bet" json:"min_bet" binding:"required"`
	MaxBet   	int32   `form:"max_bet" json:"max_bet" binding:"required"`
	SocketURL	string	`form:"socket_url" json:"socket_url" binding:"required"`
}

var router *gin.Engine

//Router Related Function
func IndexHandler(c *gin.Context){
	c.JSON(http.StatusOK, gin.H{
		"status":1,
		"message":"",
	})
}

func GetAllTable(c *gin.Context){
	session := c.MustGet("DBSession").(*mgo.Session).Copy()
	defer session.Close()

	col := session.DB("royale").C("c_table")

	var tables []Table
	err := col.Find(bson.M{}).All(&tables)
	if err != nil {
		c.JSON( http.StatusOK, gin.H{
			"status":0,
			"message":"Internal Error",
		})
		return;
	}

	c.JSON(http.StatusOK, gin.H{
		"status":1,
		"message":"",
		"tables":tables,
	})
}

//Insert data to table
func AddTable(c *gin.Context){
	session := c.MustGet("DBSession").(*mgo.Session).Copy()
	defer session.Close()

	col := session.DB("royale").C("c_table")
	errorMessage := "Fail to Insert Data";
	var table Table
	err := c.BindJSON(&table)
	if(err != nil){
		errorMessage = "Invalid Input " + err.Error()
		c.JSON(http.StatusOK, gin.H{
			"status":0,
			"message":errorMessage,
		})
		return
	}

	err = col.Insert(table)
	if err != nil {
		if mgo.IsDup(err) {
			errorMessage = "Duplicate Data"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":0,
			"message":errorMessage,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":1,
		"message":"Success Insert Data",
	})
}

//Db Related function
func EnsureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("royale").C("c_table")

	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func DBMiddleWare(s *mgo.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
        c.Set("DBSession", s)
        c.Next()
    }
}

func main() {

	//Set the database connection
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	EnsureIndex(session)

	// Use empty gin routers
	router = gin.New()
	//Use DB Middle ware 
	router.Use(DBMiddleWare(session))

	//Set the index handler
	router.GET("/", IndexHandler)
	router.GET("/tables", GetAllTable)
	router.POST("/table", AddTable)

	// Start serving the service in 
	router.Run("0.0.0.0:8225")

}