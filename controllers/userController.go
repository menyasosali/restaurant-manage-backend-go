package controllers

import "net/http"

func GetUsers(w http.ResponseWriter, r *http.Request) {

}

func GetUser(w http.ResponseWriter, r *http.Request) {

}

func SingUp(w http.ResponseWriter, r *http.Request) {

}

func Login(w http.ResponseWriter, r *http.Request) {

}

func HashPassword(password string) string {
	return ""
}

func VerifyPassword(hashPassword string, memPassword string) (bool, string) {
	return false, ""
}
