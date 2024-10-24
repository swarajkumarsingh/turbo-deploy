package userModel

import "github.com/golang-jwt/jwt/v5"

type User struct {
	Id          int    `json:"id" db:"id"`
	Username    string `json:"username" db:"username"`
	FirstName   string `json:"firstname" db:"firstname"`
	Lastname    string `json:"lastname" db:"lastname"`
	Email       string `json:"email" db:"email"`
	Password    string `json:"password" db:"password"`
	Phone       string `json:"phone" db:"phone"`
	IsActive    string `json:"is_active" db:"is_active"`
	IsDeleted   string `json:"is_deleted" db:"is_deleted"`
	Address     string `json:"address" db:"address"`
	Experience  string `json:"experience" db:"experience"`
	PrimaryGoal string `json:"primary_goal" db:"primary_goal"`
	UserRole    string `json:"user_role" db:"user_role"`
	PlanType    string `json:"plan_type" db:"plan_type"`
	Created_at  string `json:"created_on" db:"created_at"`
	Updated_at  string `json:"updated_at" db:"updated_at"`
}

type UserBody struct {
	Username    string `validate:"required" json:"username"`
	FirstName   string `validate:"required" json:"firstname"`
	LastName    string `validate:"required" json:"lastname"`
	Email       string `validate:"required" json:"email"`
	Password    string `validate:"required" json:"password"`
	Phone       string `validate:"required" json:"phone"`
	Address     string `validate:"required" json:"address"`
	Experience  string `json:"experience"`
	PrimaryGoal string `json:"primary_goal"`
	UserRole    string `json:"user_role"`
	PlanType    string `json:"plan_type"`
}

type UserUpdateBody struct {
	Username    string `validate:"required" json:"username"`
	FirstName   string `validate:"required" json:"firstname"`
	LastName    string `validate:"required" json:"lastname"`
	Address     string `validate:"required" json:"address"`
	Experience  string `validate:"required" json:"experience"`
	PrimaryGoal string `validate:"required" json:"primary_goal"`
	UserRole    string `validate:"required" json:"user_role"`
	PlanType    string `validate:"required" json:"plan_type"`
}

type Claims struct {
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}
