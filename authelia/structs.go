package authelia

type FirstFactorRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	KeepMeLoggedIn bool   `json:"keepMeLoggedIn"`
}

type StatusResponse struct {
	Status string `json:"status"`
}

type TOTPRequest struct {
	Token string `json:"token"`
}
