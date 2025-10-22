package internal

type ApplicationStatus string

const (
	AppNEW        ApplicationStatus = "NEW"        // оформлена
	AppACCEPTED   ApplicationStatus = "ACCEPTED"   // принята на склад
	AppDISPATCHED ApplicationStatus = "DISPATCHED" // отправлена
	AppDELIVERED  ApplicationStatus = "DELIVERED"  // доставлена
)
