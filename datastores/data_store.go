package datastores

import (
	"log"

	"github.com/titouanfreville/popcubeexternalapi/models"
	u "github.com/titouanfreville/popcubeexternalapi/utils"

	// Importing sql driver. They are used by gorm package and used by default from blank.
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// StoreInterface interface the Stores and usefull DB functions
type StoreInterface interface {
	Organisation() OrganisationStore
	User() UserStore
	InitConnection(user string, dbname string, password string, host string, port string) *gorm.DB
	InitDatabase(user string, dbname string, password string, host string, port string)
	CloseConnection(*gorm.DB)
}

// StoreImpl implement store interface
type StoreImpl struct{}

// Store init store
func Store() StoreInterface {
	return StoreImpl{}
}

// InitConnection init Database connection && database models
func (store StoreImpl) InitConnection(user string, dbname string, password string, host string, port string) *gorm.DB {
	connectionChain := user + ":" + password + "@(" + host + ":" + port + ")/" + dbname + "?charset=utf8&parseTime=True&loc=Local"
	db, _ := gorm.Open("mysql", connectionChain)

	// Will not set CreatedAt and LastUpdate on .Create() call
	db.Callback().Create().Remove("gorm:update_time_stamp")
	// db.Callback().Create().Remove("gorm:save_associations")

	// Will not update LastUpdate on .Save() call
	db.Callback().Update().Remove("gorm:update_time_stamp")
	// db.Callback().Update().Remove("gorm:save_associations")

	if err := db.DB().Ping(); err != nil {
		log.Print("Can't connect to database")
		log.Print(host)
		return nil
	}
	return db
}

// InitDatabase initialise a connection to the database and the database.
func (store StoreImpl) InitDatabase(user string, dbname string, password string, host string, port string) {
	db := store.InitConnection(user, dbname, password, host, port)
	db.Debug().DB().Ping()
	// Create correct tables
	db.AutoMigrate(&models.Organisation{}, &models.User{})

	// Will not set CreatedAt and LastUpdate on .Create() call
	db.Callback().Create().Remove("gorm:update_time_stamp")
	// db.Callback().Create().Remove("gorm:save_associations")

	// Will not update LastUpdate on .Save() call
	db.Callback().Update().Remove("gorm:update_time_stamp")
	// db.Callback().Update().Remove("gorm:save_associations")

	db.Debug().DB().Ping()
}

// CloseConnection close database connection
func (store StoreImpl) CloseConnection(db *gorm.DB) {
	defer db.Close()
}

/*OrganisationStore interface the organisation communication
Organisation is unique in the database. So they are no use of providing an user to get.
Delete is useless as we will down the docker stack in case an organisation leace.
*/
type OrganisationStore interface {
	Save(organisation *models.Organisation, db *gorm.DB) *u.AppError
	Update(organisation *models.Organisation, newOrganisation *models.Organisation, db *gorm.DB) *u.AppError
	Get(db *gorm.DB) []models.Organisation
	GetByID(ID uint64, db *gorm.DB) models.Organisation
	GeByName(name string, db *gorm.DB) models.Organisation
}

/*UserStore interface the user communication*/
type UserStore interface {
	Save(user *models.User, db *gorm.DB) *u.AppError
	Update(user *models.User, newUser *models.User, db *gorm.DB) *u.AppError
	GetByID(ID uint64, db *gorm.DB) models.User
	GetByUserName(userName string, db *gorm.DB) models.User
	GetByEmail(userEmail string, db *gorm.DB) models.User
	GetOrderedByDate(userDate int, db *gorm.DB) []models.User
	GetDeleted(db *gorm.DB) []models.User
	GetByNickName(nickName string, db *gorm.DB) models.User
	GetByFirstName(firstName string, db *gorm.DB) []models.User
	GetByLastName(lastName string, db *gorm.DB) []models.User
	GetByOrganisation(role *models.Organisation, db *gorm.DB) []models.User
	GetAll(db *gorm.DB) []models.User
	Delete(user *models.User, db *gorm.DB) *u.AppError
	// Login(userName string, pass string, db *gorm.DB) (models.User, *u.AppError)
}
