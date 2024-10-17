package models

type EmailDomains struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Institution string `json:"institution" gorm:"size:255;not null;unique"`
	Domain      string `json:"domain" gorm:"size:255;not null;unique"`
}

func GetEmailDomains() ([]EmailDomains, error) {

	var emailDomains []EmailDomains

	if err := DB.Find(&emailDomains).Error; err != nil {
		return emailDomains, err
	}

	return emailDomains, nil
}

func EmailDomainExists(domain string) bool {
	var count int64
	DB.Model(&EmailDomains{}).Where("domain = ?", domain).Count(&count)
	return count > 0
}
