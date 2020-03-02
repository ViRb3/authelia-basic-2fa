package authelia

var endpointVerifyUrl = "/api/verify"
var endpointFirstFactorUrl = "/api/firstfactor"
var endpointTOTPUrl = "/api/secondfactor/totp"

var (
	FirstFactorUrl string
	TOTPUrl        string
	VerifyUrl      string
)

// Builds actual Authelia URLs
func BuildUrls(baseUrl string) {
	VerifyUrl = baseUrl + endpointVerifyUrl
	FirstFactorUrl = baseUrl + endpointFirstFactorUrl
	TOTPUrl = baseUrl + endpointTOTPUrl
}
