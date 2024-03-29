package datastores

import (
	"github.com/jinzhu/gorm"
	"github.com/titouanfreville/popcubeexternalapi/models"
	u "github.com/titouanfreville/popcubeexternalapi/utils"
)

// OrganisationStoreImpl implements OrganisationSotre interface
type OrganisationStoreImpl struct{}

// Organisation Generate the struct for avatar store
func (s StoreImpl) Organisation() OrganisationStore {
	return &OrganisationStoreImpl{}
}

// Save Use to save data in BB
func (osi OrganisationStoreImpl) Save(organisation *models.Organisation, db *gorm.DB) *u.AppError {
	transaction := db.Begin()
	organisation.PreSave()
	if appError := organisation.IsValid(); appError != nil {
		transaction.Rollback()
		return u.NewLocAppError("organisationStoreImpl.Save.organisation.PreSave", appError.ID, nil, appError.DetailedError)
	}
	if !transaction.NewRecord(organisation) {
		transaction.Rollback()
		return u.NewLocAppError("organisationStoreImpl.Save", "save.transaction.create.already_exist", nil, "Organisation Name: "+organisation.OrganisationName)
	}
	if err := transaction.Create(&organisation).Error; err != nil {
		transaction.Rollback()
		return u.NewLocAppError("organisationStoreImpl.Save", "save.transaction.create.encounterError: "+err.Error(), nil, "")
	}
	transaction.Commit()
	return nil
}

// Update Used to update data in DB
func (osi OrganisationStoreImpl) Update(organisation *models.Organisation, newOrganisation *models.Organisation, db *gorm.DB) *u.AppError {

	transaction := db.Begin()
	newOrganisation.PreSave()
	if appError := organisation.IsValid(); appError != nil {
		transaction.Rollback()
		return u.NewLocAppError("organisationStoreImpl.Update.organisationOld.PreSave", appError.ID, nil, appError.DetailedError)
	}
	// if appError := newOrganisation.IsValid(); appError != nil {
	// 	transaction.Rollback()
	// 	return u.NewLocAppError("organisationStoreImpl.Update.organisationNew.PreSave", appError.ID, nil, appError.DetailedError)
	// }
	if err := transaction.Model(&organisation).Updates(&newOrganisation).Error; err != nil {
		transaction.Rollback()
		return u.NewLocAppError("organisationStoreImpl.Update", "update.transaction.updates.encounterError: "+err.Error(), nil, "")
	}
	transaction.Commit()
	return nil
}

// Get Used to get organisation from DB
func (osi OrganisationStoreImpl) Get(db *gorm.DB) []models.Organisation {
	organisation := []models.Organisation{}
	db.Find(&organisation)
	return organisation
}

// GeByName Used to get organisation from DB
func (osi OrganisationStoreImpl) GeByName(name string, db *gorm.DB) models.Organisation {
	organisation := models.EmptyOrganisation
	db.First(&organisation)
	return organisation
}

// GetByID Used to get organisation from DB
func (osi OrganisationStoreImpl) GetByID(ID uint64, db *gorm.DB) models.Organisation {
	organisation := models.EmptyOrganisation
	db.Where("idOrganisation = ?", ID).First(&organisation)
	return organisation
}
