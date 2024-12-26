package user

import (
	"context"

	"github.com/erikgmatos/gobid/internal/validator"
)

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginUserReq) Valid(tx context.Context) validator.Evaluator {
	var eval validator.Evaluator
	eval.CheckField(validator.NotBlank(req.Email), "email", "this field cannot be empty")
	eval.CheckField(validator.Matches(req.Email, validator.EmailRx), "email", "must be a valid email")
	eval.CheckField(validator.NotBlank(req.Password), "password", "this field cannot be empty")
	eval.CheckField(validator.MinChars(req.Password, 8), "password", "password must be bigger than 8 chars")

	return eval
}
