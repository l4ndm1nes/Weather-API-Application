package port

type Mailer interface {
	SendConfirmation(email, token string) error
	SendWeatherUpdate(email, city string, weatherInfo string) error
}
